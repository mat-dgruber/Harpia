package pdf

import (
	"fmt"
	"os"

	"github.com/mat-dgruber/Harpia/hrp"
)

func met_pdf_gerarDeHtml(inst hrp.Objeto, args hrp.Tupla) (hrp.Objeto, error) {
	if err := hrp.VerificaNumeroArgumentos("gerarDeHtml", false, args, 2, 2); err != nil {
		return nil, err
	}

	htmlConteudo, err := hrp.NewTexto(args[0])
	if err != nil {
		return nil, err
	}
	caminhoSaida, err := hrp.NewTexto(args[1])
	if err != nil {
		return nil, err
	}

	// Simulação / Estruturação do gerador PDF nativo
	pdfSimulado := fmt.Sprintf("%%PDF-1.4\n%% Generator: Harpia Native PDF Engine\n1 0 obj\n<< /Title (Documento Harpia) >>\nendobj\n%% HTML Content:\n%s", string(htmlConteudo.(hrp.Texto)))
	err = os.WriteFile(string(caminhoSaida.(hrp.Texto)), []byte(pdfSimulado), 0644)
	if err != nil {
		return hrp.Booleano(false), hrp.NewErroF(hrp.ValorErro, "Erro ao gravar arquivo PDF: %v", err)
	}

	return hrp.Booleano(true), nil
}

var _gerarDeHtml = hrp.NewMetodoOuPanic("gerarDeHtml", met_pdf_gerarDeHtml, "")

func init() {
	hrp.RegistraModuloImpl(&hrp.ModuloImpl{
		Info: hrp.ModuloInfo{
			Nome:    "pdf",
			Arquivo: "stdlib/pdf",
		},
		Metodos: []*hrp.Metodo{
			_gerarDeHtml,
		},
	})
}
