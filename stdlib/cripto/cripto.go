package cripto

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"

	"github.com/google/uuid"
	"github.com/natanfeitosa/portuscript/ptst"
)

// met_cripto_sha256 implementa 'sha256(texto)' -> retorna hash hexadecimal SHA256 do texto fornecido
func met_cripto_sha256(inst ptst.Objeto, args ptst.Tupla) (ptst.Objeto, error) {
	if err := ptst.VerificaNumeroArgumentos("sha256", false, args, 1, 1); err != nil {
		return nil, err
	}

	texto, err := ptst.NewTexto(args[0])
	if err != nil {
		return nil, err
	}

	hash := sha256.Sum256([]byte(texto.(ptst.Texto)))
	return ptst.Texto(hex.EncodeToString(hash[:])), nil
}

// met_cripto_codificarBase64 implementa 'codificarBase64(texto)'
func met_cripto_codificarBase64(inst ptst.Objeto, args ptst.Tupla) (ptst.Objeto, error) {
	if err := ptst.VerificaNumeroArgumentos("codificarBase64", false, args, 1, 1); err != nil {
		return nil, err
	}

	texto, err := ptst.NewTexto(args[0])
	if err != nil {
		return nil, err
	}

	res := base64.StdEncoding.EncodeToString([]byte(texto.(ptst.Texto)))
	return ptst.Texto(res), nil
}

// met_cripto_decodificarBase64 implementa 'decodificarBase64(texto)'
func met_cripto_decodificarBase64(inst ptst.Objeto, args ptst.Tupla) (ptst.Objeto, error) {
	if err := ptst.VerificaNumeroArgumentos("decodificarBase64", false, args, 1, 1); err != nil {
		return nil, err
	}

	texto, err := ptst.NewTexto(args[0])
	if err != nil {
		return nil, err
	}

	bytes, err := base64.StdEncoding.DecodeString(string(texto.(ptst.Texto)))
	if err != nil {
		return nil, ptst.NewErroF(ptst.ValorErro, "Erro ao decodificar string Base64: %v", err)
	}

	return ptst.Texto(bytes), nil
}

// met_cripto_uuid implementa 'uuid()' -> gera string UUID v4 aleatório e robusto
func met_cripto_uuid(inst ptst.Objeto, args ptst.Tupla) (ptst.Objeto, error) {
	if err := ptst.VerificaNumeroArgumentos("uuid", false, args, 0, 0); err != nil {
		return nil, err
	}

	id := uuid.New().String()
	return ptst.Texto(id), nil
}

var _sha256 = ptst.NewMetodoOuPanic("sha256", met_cripto_sha256, "")
var _codificarBase64 = ptst.NewMetodoOuPanic("codificarBase64", met_cripto_codificarBase64, "")
var _decodificarBase64 = ptst.NewMetodoOuPanic("decodificarBase64", met_cripto_decodificarBase64, "")
var _uuid = ptst.NewMetodoOuPanic("uuid", met_cripto_uuid, "")

func init() {
	ptst.RegistraModuloImpl(&ptst.ModuloImpl{
		Info: ptst.ModuloInfo{
			Nome:    "cripto",
			Arquivo: "stdlib/cripto",
		},
		Metodos: []*ptst.Metodo{
			_sha256,
			_codificarBase64,
			_decodificarBase64,
			_uuid,
		},
	})
}
