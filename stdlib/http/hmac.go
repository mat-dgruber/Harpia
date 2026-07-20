package http

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"

	"github.com/mat-dgruber/Harpia/hrp"
)

func met_assinar_hmac(_ hrp.Objeto, args hrp.Tupla) (hrp.Objeto, error) {
	if err := hrp.VerificaNumeroArgumentos("assinar_hmac", false, args, 2, 2); err != nil {
		return nil, err
	}
	chave, err := hrp.NewTexto(args[0])
	if err != nil {
		return nil, err
	}
	mensagem, err := hrp.NewTexto(args[1])
	if err != nil {
		return nil, err
	}

	h := hmac.New(sha256.New, []byte(chave.(hrp.Texto)))
	h.Write([]byte(mensagem.(hrp.Texto)))
	mac := h.Sum(nil)
	return hrp.Texto(hex.EncodeToString(mac)), nil
}

func met_verificar_hmac(_ hrp.Objeto, args hrp.Tupla) (hrp.Objeto, error) {
	if err := hrp.VerificaNumeroArgumentos("verificar_hmac", false, args, 3, 3); err != nil {
		return nil, err
	}
	chave, err := hrp.NewTexto(args[0])
	if err != nil {
		return nil, err
	}
	mensagem, err := hrp.NewTexto(args[1])
	if err != nil {
		return nil, err
	}
	assinatura, err := hrp.NewTexto(args[2])
	if err != nil {
		return nil, err
	}

	h := hmac.New(sha256.New, []byte(chave.(hrp.Texto)))
	h.Write([]byte(mensagem.(hrp.Texto)))
	macEsperado := h.Sum(nil)
	macRecebido, errHex := hex.DecodeString(string(assinatura.(hrp.Texto)))
	if errHex != nil {
		return hrp.Falso, nil
	}

	if hmac.Equal(macEsperado, macRecebido) {
		return hrp.Verdadeiro, nil
	}
	return hrp.Falso, nil
}
