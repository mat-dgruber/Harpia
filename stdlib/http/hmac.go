package http

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"

	"github.com/mat-dgruber/Harpia/ptst"
)

func met_assinar_hmac(_ ptst.Objeto, args ptst.Tupla) (ptst.Objeto, error) {
	if err := ptst.VerificaNumeroArgumentos("assinar_hmac", false, args, 2, 2); err != nil {
		return nil, err
	}
	chave, err := ptst.NewTexto(args[0])
	if err != nil {
		return nil, err
	}
	mensagem, err := ptst.NewTexto(args[1])
	if err != nil {
		return nil, err
	}

	h := hmac.New(sha256.New, []byte(chave.(ptst.Texto)))
	h.Write([]byte(mensagem.(ptst.Texto)))
	mac := h.Sum(nil)
	return ptst.Texto(hex.EncodeToString(mac)), nil
}

func met_verificar_hmac(_ ptst.Objeto, args ptst.Tupla) (ptst.Objeto, error) {
	if err := ptst.VerificaNumeroArgumentos("verificar_hmac", false, args, 3, 3); err != nil {
		return nil, err
	}
	chave, err := ptst.NewTexto(args[0])
	if err != nil {
		return nil, err
	}
	mensagem, err := ptst.NewTexto(args[1])
	if err != nil {
		return nil, err
	}
	assinatura, err := ptst.NewTexto(args[2])
	if err != nil {
		return nil, err
	}

	h := hmac.New(sha256.New, []byte(chave.(ptst.Texto)))
	h.Write([]byte(mensagem.(ptst.Texto)))
	macEsperado := h.Sum(nil)
	macRecebido, errHex := hex.DecodeString(string(assinatura.(ptst.Texto)))
	if errHex != nil {
		return ptst.Falso, nil
	}

	if hmac.Equal(macEsperado, macRecebido) {
		return ptst.Verdadeiro, nil
	}
	return ptst.Falso, nil
}
