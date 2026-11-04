// Package cripto implementa rotinas de criptografia simétrica (AES-256-GCM),
// codificação Base64, geração de hash SHA-256 e identificadores UUID v4 aleatórios.
package cripto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"io"
	"strings"

	"github.com/google/uuid"
	"github.com/mat-dgruber/Harpia/hrp"
	"golang.org/x/crypto/scrypt"
)

// met_cripto_sha256 implementa 'sha256(texto)' em nível de script Harpia.
// Gera um digest/hash hexadecimal unidirecional SHA-256 determinístico.
func met_cripto_sha256(inst hrp.Objeto, args hrp.Tupla) (hrp.Objeto, error) {
	if err := hrp.VerificaNumeroArgumentos("sha256", false, args, 1, 1); err != nil {
		return nil, err
	}

	texto, err := hrp.NewTexto(args[0])
	if err != nil {
		return nil, err
	}

	hash := sha256.Sum256([]byte(texto.(hrp.Texto)))
	return hrp.Texto(hex.EncodeToString(hash[:])), nil
}

// met_cripto_codificarBase64 implementa 'codificarBase64(texto)' em nível de script Harpia.
// Codifica dados em formato Base64 padrão seguro para transporte de payloads binários via rede.
func met_cripto_codificarBase64(inst hrp.Objeto, args hrp.Tupla) (hrp.Objeto, error) {
	if err := hrp.VerificaNumeroArgumentos("codificarBase64", false, args, 1, 1); err != nil {
		return nil, err
	}

	texto, err := hrp.NewTexto(args[0])
	if err != nil {
		return nil, err
	}

	res := base64.StdEncoding.EncodeToString([]byte(texto.(hrp.Texto)))
	return hrp.Texto(res), nil
}

// met_cripto_decodificarBase64 implementa 'decodificarBase64(texto)' em nível de script Harpia.
// Decodifica dados em formato Base64 de volta para sua representação original de string.
func met_cripto_decodificarBase64(inst hrp.Objeto, args hrp.Tupla) (hrp.Objeto, error) {
	if err := hrp.VerificaNumeroArgumentos("decodificarBase64", false, args, 1, 1); err != nil {
		return nil, err
	}

	texto, err := hrp.NewTexto(args[0])
	if err != nil {
		return nil, err
	}

	bytes, err := base64.StdEncoding.DecodeString(string(texto.(hrp.Texto)))
	if err != nil {
		return nil, hrp.NewErroF(hrp.ValorErro, "Erro ao decodificar string Base64: %v", err)
	}

	return hrp.Texto(bytes), nil
}

// met_cripto_uuid implementa 'uuid()' em nível de script Harpia.
// Gera um identificador único de 128 bits aleatório (UUID v4) conforme a RFC 4122.
func met_cripto_uuid(inst hrp.Objeto, args hrp.Tupla) (hrp.Objeto, error) {
	if err := hrp.VerificaNumeroArgumentos("uuid", false, args, 0, 0); err != nil {
		return nil, err
	}

	id := uuid.New().String()
	return hrp.Texto(id), nil
}

// met_cripto_cifrar implementa 'cifrar(texto, chave)' em nível de script Harpia.
// Cifra dados utilizando o algoritmo simétrico forte AES-256-GCM (Criptografia de Envelope).
// O hash SHA-256 da chave fornecida é usado como a chave real de cifragem de 256 bits.
func met_cripto_cifrar(inst hrp.Objeto, args hrp.Tupla) (hrp.Objeto, error) {
	if err := hrp.VerificaNumeroArgumentos("cifrar", false, args, 2, 2); err != nil {
		return nil, err
	}
	texto, err := hrp.NewTexto(args[0])
	if err != nil {
		return nil, err
	}
	chaveStr, err := hrp.NewTexto(args[1])
	if err != nil {
		return nil, err
	}

	sal := make([]byte, tamanhoSal)
	if _, err := io.ReadFull(rand.Reader, sal); err != nil {
		return nil, hrp.NewErroF(hrp.ValorErro, "Erro ao gerar sal: %v", err)
	}
	chave, err := derivarChaveV2(string(chaveStr.(hrp.Texto)), sal)
	if err != nil {
		return nil, err
	}
	gcm, err := novoGCM(chave)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, hrp.NewErroF(hrp.ValorErro, "Erro ao gerar nonce: %v", err)
	}

	// envelope v2: "v2:" + base64(sal || nonce || texto cifrado)
	envelope := append(append(sal, nonce...), gcm.Seal(nil, nonce, []byte(texto.(hrp.Texto)), nil)...)
	return hrp.Texto(prefixoV2 + base64.StdEncoding.EncodeToString(envelope)), nil
}

const (
	prefixoV2  = "v2:"
	tamanhoSal = 16
)

// derivarChaveV2 deriva a chave AES-256 da senha com scrypt (N=2^15, r=8, p=1), lento de propósito
// contra força bruta. O formato legado usava sha256(senha), sem sal.
func derivarChaveV2(senha string, sal []byte) ([]byte, error) {
	chave, err := scrypt.Key([]byte(senha), sal, 1<<15, 8, 1, 32)
	if err != nil {
		return nil, hrp.NewErroF(hrp.ValorErro, "Erro ao derivar chave: %v", err)
	}
	return chave, nil
}

func novoGCM(chave []byte) (cipher.AEAD, error) {
	block, err := aes.NewCipher(chave)
	if err != nil {
		return nil, hrp.NewErroF(hrp.ValorErro, "Erro ao inicializar cifra AES: %v", err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, hrp.NewErroF(hrp.ValorErro, "Erro ao inicializar GCM: %v", err)
	}
	return gcm, nil
}

// met_cripto_decifrar implementa 'decifrar(textoCifrado, chave)' em nível de script Harpia.
// Decifra dados em formato Base64 utilizando o algoritmo simétrico AES-256-GCM.
func met_cripto_decifrar(inst hrp.Objeto, args hrp.Tupla) (hrp.Objeto, error) {
	if err := hrp.VerificaNumeroArgumentos("decifrar", false, args, 2, 2); err != nil {
		return nil, err
	}
	cifradoB64, err := hrp.NewTexto(args[0])
	if err != nil {
		return nil, err
	}
	chaveStr, err := hrp.NewTexto(args[1])
	if err != nil {
		return nil, err
	}

	entrada := string(cifradoB64.(hrp.Texto))
	v2 := strings.HasPrefix(entrada, prefixoV2)
	dados, err := base64.StdEncoding.DecodeString(strings.TrimPrefix(entrada, prefixoV2))
	if err != nil {
		return nil, hrp.NewErroF(hrp.ValorErro, "Erro ao decodificar Base64: %v", err)
	}

	var chave []byte
	if v2 {
		if len(dados) < tamanhoSal {
			return nil, hrp.NewErroF(hrp.ValorErro, "Dados cifrados inválidos (sem sal)")
		}
		if chave, err = derivarChaveV2(string(chaveStr.(hrp.Texto)), dados[:tamanhoSal]); err != nil {
			return nil, err
		}
		dados = dados[tamanhoSal:]
	} else {
		legado := sha256.Sum256([]byte(chaveStr.(hrp.Texto))) // formato antigo, só para leitura
		chave = legado[:]
	}
	gcm, err := novoGCM(chave)
	if err != nil {
		return nil, err
	}

	nonceSize := gcm.NonceSize()
	if len(dados) < nonceSize {
		return nil, hrp.NewErroF(hrp.ValorErro, "Dados cifrados inválidos (tamanho inferior ao nonce)")
	}

	nonce, ciphertext := dados[:nonceSize], dados[nonceSize:]
	plano, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, hrp.NewErroF(hrp.ValorErro, "Erro ao decifrar dados (chave incorreta ou corrompido): %v", err)
	}

	return hrp.Texto(plano), nil
}

var _sha256 = hrp.NewMetodoOuPanic("sha256", met_cripto_sha256, "Gera o hash SHA-256 hexadecimal de um texto.")
var _codificarBase64 = hrp.NewMetodoOuPanic("codificarBase64", met_cripto_codificarBase64, "Codifica uma string no formato Base64.")
var _decodificarBase64 = hrp.NewMetodoOuPanic("decodificarBase64", met_cripto_decodificarBase64, "Decodifica uma string Base64 para texto simples.")
var _uuid = hrp.NewMetodoOuPanic("uuid", met_cripto_uuid, "Gera um UUID v4 aleatório forte.")
var _cifrar = hrp.NewMetodoOuPanic("cifrar", met_cripto_cifrar, "Cifra um texto plano usando AES-256-GCM com a chave fornecida.")
var _decifrar = hrp.NewMetodoOuPanic("decifrar", met_cripto_decifrar, "Decifra um payload encriptado com AES-256-GCM utilizando a chave correspondente.")

func init() {
	// Registra o módulo 'cripto' na biblioteca padrão do Harpia.
	hrp.RegistraModuloImpl(&hrp.ModuloImpl{
		Info: hrp.ModuloInfo{
			Nome:    "cripto",
			Arquivo: "stdlib/cripto",
		},
		Metodos: []*hrp.Metodo{
			_sha256,
			_codificarBase64,
			_decodificarBase64,
			_uuid,
			_cifrar,
			_decifrar,
		},
	})
}
