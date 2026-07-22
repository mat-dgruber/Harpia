package autenticacao

import (
	"crypto/hmac"
	"crypto/sha256"

	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/mat-dgruber/Harpia/hrp"
	"golang.org/x/crypto/bcrypt"
)

// met_gerarHashSenha implementa 'gerarHashSenha(senha)'
func met_gerarHashSenha(inst hrp.Objeto, args hrp.Tupla) (hrp.Objeto, error) {
	if err := hrp.VerificaNumeroArgumentos("gerarHashSenha", false, args, 1, 1); err != nil {
		return nil, err
	}
	senha, err := hrp.NewTexto(args[0])
	if err != nil {
		return nil, err
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(senha.(hrp.Texto)), bcrypt.DefaultCost)
	if err != nil {
		return nil, hrp.NewErroF(hrp.ValorErro, "Erro ao gerar hash de senha: %v", err)
	}

	return hrp.Texto(string(hash)), nil
}

// met_verificarSenha implementa 'verificarSenha(senha, hash)'
func met_verificarSenha(inst hrp.Objeto, args hrp.Tupla) (hrp.Objeto, error) {
	if err := hrp.VerificaNumeroArgumentos("verificarSenha", false, args, 2, 2); err != nil {
		return nil, err
	}
	senha, err := hrp.NewTexto(args[0])
	if err != nil {
		return nil, err
	}
	hash, err := hrp.NewTexto(args[1])
	if err != nil {
		return nil, err
	}

	err = bcrypt.CompareHashAndPassword([]byte(hash.(hrp.Texto)), []byte(senha.(hrp.Texto)))
	return hrp.Booleano(err == nil), nil
}

// met_criarJwt implementa 'criarJwt(payloadMapa, segredo)'
func met_criarJwt(inst hrp.Objeto, args hrp.Tupla) (hrp.Objeto, error) {
	if err := hrp.VerificaNumeroArgumentos("criarJwt", false, args, 2, 2); err != nil {
		return nil, err
	}

	mapa, ok := args[0].(hrp.Mapa)
	if !ok {
		return nil, hrp.NewErroF(hrp.TipoErro, "O primeiro argumento de 'criarJwt' deve ser um Mapa")
	}

	segredo, err := hrp.NewTexto(args[1])
	if err != nil {
		return nil, err
	}

	headerJSON := `{"alg":"HS256","typ":"JWT"}`
	headerB64 := base64.RawURLEncoding.EncodeToString([]byte(headerJSON))

	payloadMap := make(map[string]interface{})
	for k, v := range mapa {
		payloadMap[k] = fmt.Sprintf("%v", v)
	}


	if _, ok := payloadMap["exp"]; !ok {
		payloadMap["exp"] = time.Now().Add(24 * time.Hour).Unix()
	}

	payloadJSON, _ := json.Marshal(payloadMap)
	payloadB64 := base64.RawURLEncoding.EncodeToString(payloadJSON)

	unsignedToken := fmt.Sprintf("%s.%s", headerB64, payloadB64)

	h := hmac.New(sha256.New, []byte(segredo.(hrp.Texto)))
	h.Write([]byte(unsignedToken))
	signatureB64 := base64.RawURLEncoding.EncodeToString(h.Sum(nil))

	return hrp.Texto(fmt.Sprintf("%s.%s", unsignedToken, signatureB64)), nil
}

// met_validarJwt implementa 'validarJwt(token, segredo)'
func met_validarJwt(inst hrp.Objeto, args hrp.Tupla) (hrp.Objeto, error) {
	if err := hrp.VerificaNumeroArgumentos("validarJwt", false, args, 2, 2); err != nil {
		return nil, err
	}

	tokenStr, err := hrp.NewTexto(args[0])
	if err != nil {
		return nil, err
	}
	segredoStr, err := hrp.NewTexto(args[1])
	if err != nil {
		return nil, err
	}

	partes := strings.Split(string(tokenStr.(hrp.Texto)), ".")
	if len(partes) != 3 {
		return hrp.Booleano(false), nil
	}

	unsignedToken := fmt.Sprintf("%s.%s", partes[0], partes[1])
	h := hmac.New(sha256.New, []byte(segredoStr.(hrp.Texto)))
	h.Write([]byte(unsignedToken))
	expectedSig := base64.RawURLEncoding.EncodeToString(h.Sum(nil))

	if hmac.Equal([]byte(partes[2]), []byte(expectedSig)) {
		return hrp.Booleano(true), nil
	}

	return hrp.Booleano(false), nil
}

var _gerarHashSenha = hrp.NewMetodoOuPanic("gerarHashSenha", met_gerarHashSenha, "")
var _verificarSenha = hrp.NewMetodoOuPanic("verificarSenha", met_verificarSenha, "")
var _criarJwt = hrp.NewMetodoOuPanic("criarJwt", met_criarJwt, "")
var _validarJwt = hrp.NewMetodoOuPanic("validarJwt", met_validarJwt, "")

func init() {
	hrp.RegistraModuloImpl(&hrp.ModuloImpl{
		Info: hrp.ModuloInfo{
			Nome:    "autenticacao",
			Arquivo: "stdlib/autenticacao",
		},
		Metodos: []*hrp.Metodo{
			_gerarHashSenha,
			_verificarSenha,
			_criarJwt,
			_validarJwt,
		},
	})
}
