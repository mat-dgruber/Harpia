package logs

import (
	"bytes"
	"os"
	"strings"
	"testing"

	"github.com/mat-dgruber/Harpia/hrp"
)

func TestLogsOutput(t *testing.T) {
	// Desativa cores para teste limpo
	met_logs_configurar(nil, hrp.Tupla{hrp.Texto("texto"), hrp.Booleano(false)})

	// Captura stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	args := hrp.Tupla{hrp.Texto("Conexão estabelecida"), hrp.Mapa{"ip": hrp.Texto("10.0.0.1")}}
	_, err := met_logs_info(nil, args)

	w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Fatalf("Erro ao chamar log info: %v", err)
	}

	var buf bytes.Buffer
	_, _ = buf.ReadFrom(r)
	output := buf.String()

	if !strings.Contains(output, "INFO: Conexão estabelecida") {
		t.Errorf("Saída de log incorreta, obtive:\n%s", output)
	}
	if !strings.Contains(output, `"ip":"10.0.0.1"`) {
		t.Errorf("Saída de metadados incorreta, obtive:\n%s", output)
	}
}

func TestLogsConfigurarJSON(t *testing.T) {
	// Muda para JSON
	_, errConf := met_logs_configurar(nil, hrp.Tupla{hrp.Texto("json")})
	if errConf != nil {
		t.Fatalf("Erro ao configurar logs: %v", errConf)
	}
	defer func() {
		// Restaura para texto
		met_logs_configurar(nil, hrp.Tupla{hrp.Texto("texto")})
	}()

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	args := hrp.Tupla{hrp.Texto("Acesso negado")}
	_, err := met_logs_erro(nil, args)

	w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Fatalf("Erro ao logar erro: %v", err)
	}

	var buf bytes.Buffer
	_, _ = buf.ReadFrom(r)
	output := buf.String()

	if !strings.Contains(output, `"nivel":"erro"`) || !strings.Contains(output, `"msg":"Acesso negado"`) {
		t.Errorf("Saída de log JSON incorreta, obtive:\n%s", output)
	}
}
