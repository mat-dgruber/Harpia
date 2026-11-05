package cmd

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestChecarVersaoMock(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`[{"name": "v1.5.0"}]`))
	}))
	defer ts.Close()

	versao, err := obterUltimaVersaoURL(ts.URL)
	if err != nil {
		t.Fatalf("erro ao obter versao: %v", err)
	}
	if versao != "1.5.0" {
		t.Errorf("esperava 1.5.0, obteve %s", versao)
	}
}

func TestComandoAtualizarComandaOuAlias(t *testing.T) {
	cmd := comandoAtualizar()
	if cmd.Use != "atualizar" {
		t.Errorf("esperava Use 'atualizar', obteve '%s'", cmd.Use)
	}

	temAlias := false
	for _, alias := range cmd.Aliases {
		if alias == "atualize" {
			temAlias = true
			break
		}
	}
	if !temAlias {
		t.Errorf("esperava alias 'atualize' no comando atualizar")
	}
}

func TestComandoAtualizarChecarFlag(t *testing.T) {
	cmd := comandoAtualizar()
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	flag := cmd.Flags().Lookup("checar")
	if flag == nil {
		t.Fatalf("flag --checar nao encontrada")
	}

	if err := cmd.ParseFlags([]string{"--checar"}); err != nil {
		t.Fatalf("erro ao processar flag --checar: %v", err)
	}

	val, err := cmd.Flags().GetBool("checar")
	if err != nil {
		t.Fatalf("erro ao obter valor da flag checar: %v", err)
	}
	if !val {
		t.Errorf("esperava flag checar como true apos parse de '--checar', obteve false")
	}
}

// M18: a versão remota é a maior semver (a API de tags não ordena por versão) e o comando não
// anuncia uma atualização automática que não existe.
func TestAtualizarEscolheMaiorSemverEMostraLink(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`[{"name": "v1.2.0"}, {"name": "nightly"}, {"name": "v1.10.0"}, {"name": "v1.9.0"}]`))
	}))
	defer ts.Close()

	versao, err := obterUltimaVersaoURL(ts.URL)
	if err != nil || versao != "1.10.0" {
		t.Fatalf("esperava 1.10.0, veio %q (%v)", versao, err)
	}

	antigaURL, antigaVersao := urlTagsAtualizar, Version
	urlTagsAtualizar, Version = ts.URL, "1.0.0"
	t.Cleanup(func() { urlTagsAtualizar, Version = antigaURL, antigaVersao })
	c := comandoAtualizar()
	var out bytes.Buffer
	c.SetOut(&out)
	c.SetErr(&out)
	c.SetArgs([]string{})
	if err := c.Execute(); err != nil {
		t.Fatal(err)
	}
	if strings.Contains(out.String(), "Iniciando atualização") || !strings.Contains(out.String(), "releases/tag/v1.10.0") {
		t.Fatalf("saída inesperada: %q", out.String())
	}
}
