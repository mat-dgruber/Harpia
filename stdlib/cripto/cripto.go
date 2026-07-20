package cripto

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"

	"github.com/google/uuid"
	"github.com/mat-dgruber/Harpia/hrp"
)

// met_cripto_sha256 implementa 'sha256(texto)' -> retorna hash hexadecimal SHA256 do texto fornecido
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

// met_cripto_codificarBase64 implementa 'codificarBase64(texto)'
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

// met_cripto_decodificarBase64 implementa 'decodificarBase64(texto)'
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

// met_cripto_uuid implementa 'uuid()' -> gera string UUID v4 aleatório e robusto
func met_cripto_uuid(inst hrp.Objeto, args hrp.Tupla) (hrp.Objeto, error) {
	if err := hrp.VerificaNumeroArgumentos("uuid", false, args, 0, 0); err != nil {
		return nil, err
	}

	id := uuid.New().String()
	return hrp.Texto(id), nil
}

var _sha256 = hrp.NewMetodoOuPanic("sha256", met_cripto_sha256, "")
var _codificarBase64 = hrp.NewMetodoOuPanic("codificarBase64", met_cripto_codificarBase64, "")
var _decodificarBase64 = hrp.NewMetodoOuPanic("decodificarBase64", met_cripto_decodificarBase64, "")
var _uuid = hrp.NewMetodoOuPanic("uuid", met_cripto_uuid, "")

func init() {
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
		},
	})
}
