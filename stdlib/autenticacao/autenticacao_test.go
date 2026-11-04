package autenticacao

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/mat-dgruber/Harpia/hrp"
)

func TestGerarEVerificarSenha(t *testing.T) {
	senha := hrp.Texto("SenhaSegura123!")
	resHash, err := met_gerarHashSenha(nil, hrp.Tupla{senha})
	if err != nil {
		t.Fatalf("erro ao gerar hash: %v", err)
	}

	hashStr, ok := resHash.(hrp.Texto)
	if !ok || len(hashStr) == 0 {
		t.Fatalf("hash invalido retornado")
	}

	// Senha correta
	resVerif, err := met_verificarSenha(nil, hrp.Tupla{senha, hashStr})
	if err != nil {
		t.Fatalf("erro ao verificar senha: %v", err)
	}
	if bool(resVerif.(hrp.Booleano)) != true {
		t.Errorf("esperava senha valida")
	}

	// Senha incorreta
	resVerifIncorreta, err := met_verificarSenha(nil, hrp.Tupla{hrp.Texto("Errada!"), hashStr})
	if err != nil {
		t.Fatalf("erro ao verificar senha incorreta: %v", err)
	}
	if bool(resVerifIncorreta.(hrp.Booleano)) != false {
		t.Errorf("esperava senha invalida")
	}
}

func TestCriarEValidarJWT(t *testing.T) {
	mapa := hrp.Mapa{
		"usuario": hrp.Texto("admin"),
		"role":    hrp.Texto("super"),
	}
	segredo := hrp.Texto("chave_secreta_super_segura_123")

	resJwt, err := met_criarJwt(nil, hrp.Tupla{mapa, segredo})
	if err != nil {
		t.Fatalf("erro ao criar JWT: %v", err)
	}

	jwtStr := resJwt.(hrp.Texto)

	// Validar com segredo correto
	resValido, err := met_validarJwt(nil, hrp.Tupla{jwtStr, segredo})
	if err != nil {
		t.Fatalf("erro ao validar JWT: %v", err)
	}
	if bool(resValido.(hrp.Booleano)) != true {
		t.Errorf("esperava JWT valido")
	}

	// Validar com segredo incorreto
	resInvalido, err := met_validarJwt(nil, hrp.Tupla{jwtStr, hrp.Texto("segredo_errado")})
	if err != nil {
		t.Fatalf("erro ao validar JWT com segredo errado: %v", err)
	}
	if bool(resInvalido.(hrp.Booleano)) != false {
		t.Errorf("esperava JWT invalido para segredo incorreto")
	}
}

func TestJWTExpirado(t *testing.T) {
	expiradoUnix := time.Now().Add(-10 * time.Minute).Unix()
	mapaExpirado := hrp.Mapa{
		"usuario": hrp.Texto("usuario_expirado"),
		"exp":     hrp.Inteiro(expiradoUnix),
	}
	segredo := hrp.Texto("segredo_123")

	resJwt, err := met_criarJwt(nil, hrp.Tupla{mapaExpirado, segredo})
	if err != nil {
		t.Fatalf("erro ao criar JWT expirado: %v", err)
	}

	jwtStr := resJwt.(hrp.Texto)

	resValido, err := met_validarJwt(nil, hrp.Tupla{jwtStr, segredo})
	if err != nil {
		t.Fatalf("erro ao validar JWT: %v", err)
	}
	if bool(resValido.(hrp.Booleano)) != false {
		t.Errorf("esperava que JWT expirado fosse recusado (retornando falso)")
	}
}

func TestJWT_RejeitarAlgNone(t *testing.T) {
	segredo := hrp.Texto("segredo_secreto")
	headerNone := base64.RawURLEncoding.EncodeToString([]byte(`{"alg":"none","typ":"JWT"}`))
	payload := base64.RawURLEncoding.EncodeToString([]byte(`{"sub":"123","exp":9999999999}`))
	unsignedToken := fmt.Sprintf("%s.%s", headerNone, payload)

	h := hmac.New(sha256.New, []byte(string(segredo)))
	h.Write([]byte(unsignedToken))
	sig := base64.RawURLEncoding.EncodeToString(h.Sum(nil))

	token := hrp.Texto(fmt.Sprintf("%s.%s", unsignedToken, sig))

	valido, err := met_validarJwt(nil, hrp.Tupla{token, segredo})
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if bool(valido.(hrp.Booleano)) != false {
		t.Errorf("esperava falso para header com alg:none")
	}
}

func TestJWT_PayloadInvalido(t *testing.T) {
	segredo := hrp.Texto("segredo_secreto")
	header := base64.RawURLEncoding.EncodeToString([]byte(`{"alg":"HS256","typ":"JWT"}`))
	payloadInvalido := base64.RawURLEncoding.EncodeToString([]byte(`{este_nao_e_um_json_valido`))
	unsignedToken := fmt.Sprintf("%s.%s", header, payloadInvalido)

	h := hmac.New(sha256.New, []byte(string(segredo)))
	h.Write([]byte(unsignedToken))
	sig := base64.RawURLEncoding.EncodeToString(h.Sum(nil))

	token := hrp.Texto(fmt.Sprintf("%s.%s", unsignedToken, sig))

	valido, err := met_validarJwt(nil, hrp.Tupla{token, segredo})
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if bool(valido.(hrp.Booleano)) != false {
		t.Errorf("esperava falso para payload que nao e JSON valido")
	}
}

func TestJWT_ExpNoFuturoENoPassado(t *testing.T) {
	segredo := hrp.Texto("segredo_secreto")

	// Futuro
	expFuturo := time.Now().Add(1 * time.Hour).Unix()
	mapaFuturo := hrp.Mapa{
		"sub": hrp.Texto("user1"),
		"exp": hrp.Inteiro(expFuturo),
	}
	resJwt, err := met_criarJwt(nil, hrp.Tupla{mapaFuturo, segredo})
	if err != nil {
		t.Fatalf("erro ao criar jwt: %v", err)
	}
	valido, err := met_validarJwt(nil, hrp.Tupla{resJwt, segredo})
	if err != nil {
		t.Fatalf("erro ao validar jwt: %v", err)
	}
	if bool(valido.(hrp.Booleano)) != true {
		t.Errorf("esperava verdadeiro para exp no futuro")
	}

	// Passado
	expPassado := time.Now().Add(-1 * time.Hour).Unix()
	mapaPassado := hrp.Mapa{
		"sub": hrp.Texto("user1"),
		"exp": hrp.Inteiro(expPassado),
	}
	resJwtPassado, err := met_criarJwt(nil, hrp.Tupla{mapaPassado, segredo})
	if err != nil {
		t.Fatalf("erro ao criar jwt: %v", err)
	}
	validoPassado, err := met_validarJwt(nil, hrp.Tupla{resJwtPassado, segredo})
	if err != nil {
		t.Fatalf("erro ao validar jwt: %v", err)
	}
	if bool(validoPassado.(hrp.Booleano)) != false {
		t.Errorf("esperava falso para exp no passado")
	}

	// exp não numérico (ex: string não numérica)
	mapaInvalido := hrp.Mapa{
		"sub": hrp.Texto("user1"),
		"exp": hrp.Texto("nao-e-numero"),
	}
	resJwtInvalido, err := met_criarJwt(nil, hrp.Tupla{mapaInvalido, segredo})
	if err != nil {
		t.Fatalf("erro ao criar jwt: %v", err)
	}
	validoInvalido, err := met_validarJwt(nil, hrp.Tupla{resJwtInvalido, segredo})
	if err != nil {
		t.Fatalf("erro ao validar jwt: %v", err)
	}
	if bool(validoInvalido.(hrp.Booleano)) != false {
		t.Errorf("esperava falso para exp nao numerico")
	}
}

func TestJWT_PreservarTiposNativosPayload(t *testing.T) {
	mapa := hrp.Mapa{
		"idade": hrp.Inteiro(42),
		"score": hrp.Decimal(9.75),
		"ativo": hrp.Booleano(true),
		"nome":  hrp.Texto("João"),
	}
	segredo := hrp.Texto("segredo_secreto")

	resJwt, err := met_criarJwt(nil, hrp.Tupla{mapa, segredo})
	if err != nil {
		t.Fatalf("erro ao criar jwt: %v", err)
	}

	jwtStr := string(resJwt.(hrp.Texto))
	partes := strings.Split(jwtStr, ".")
	if len(partes) != 3 {
		t.Fatalf("jwt invalido: %s", jwtStr)
	}

	payloadBytes, err := base64.RawURLEncoding.DecodeString(partes[1])
	if err != nil {
		t.Fatalf("erro ao decodificar payload: %v", err)
	}

	var payloadMap map[string]interface{}
	if err := json.Unmarshal(payloadBytes, &payloadMap); err != nil {
		t.Fatalf("erro ao parsear payload json: %v", err)
	}

	if _, ok := payloadMap["idade"].(float64); !ok {
		t.Errorf("esperava que 'idade' fosse float64/numero no json, obteve: %T (%v)", payloadMap["idade"], payloadMap["idade"])
	}
	if _, ok := payloadMap["score"].(float64); !ok {
		t.Errorf("esperava que 'score' fosse float64/numero no json, obteve: %T (%v)", payloadMap["score"], payloadMap["score"])
	}
	if _, ok := payloadMap["ativo"].(bool); !ok {
		t.Errorf("esperava que 'ativo' fosse bool no json, obteve: %T (%v)", payloadMap["ativo"], payloadMap["ativo"])
	}
	if _, ok := payloadMap["nome"].(string); !ok {
		t.Errorf("esperava que 'nome' fosse string no json, obteve: %T (%v)", payloadMap["nome"], payloadMap["nome"])
	}
}
