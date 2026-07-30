package autenticacao

import (
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
