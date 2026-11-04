package cmd

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

// TestCoerrenciaRuntimeWebHeader assevera que o header de import gerado por
// GerarImportHeaderWeb está em sincronia com os exports reais de runtime-web.js
// e com FallbackRuntimeWebJS, sem identificadores duplicados.
func TestCoerrenciaRuntimeWebHeader(t *testing.T) {
	extrairExports := func(content string) map[string]bool {
		lines := strings.Split(content, "\n")
		exports := make(map[string]bool)
		for _, line := range lines {
			line = strings.TrimSpace(line)
			var name string
			if strings.HasPrefix(line, "export async function ") {
				parts := strings.Fields(line)
				if len(parts) >= 4 {
					name = parts[3]
				}
			} else if strings.HasPrefix(line, "export function ") ||
				strings.HasPrefix(line, "export const ") ||
				strings.HasPrefix(line, "export let ") ||
				strings.HasPrefix(line, "export var ") ||
				strings.HasPrefix(line, "export class ") {
				parts := strings.Fields(line)
				if len(parts) >= 3 {
					name = parts[2]
				}
			}
			if name != "" {
				if idx := strings.IndexAny(name, "(="); idx >= 0 {
					name = name[:idx]
				}
				name = strings.TrimSpace(name)
				if name != "" {
					exports[name] = true
				}
			}
		}
		return exports
	}

	// 1. Ler os exports reais do runtime-web.js
	runtimeWebPath := filepath.Join("..", "stdlib", "web", "runtime-web.js")
	raw, err := os.ReadFile(runtimeWebPath)
	if err != nil {
		t.Skipf("runtime-web.js não encontrado: %v", err)
	}

	actualExports := extrairExports(string(raw))

	// 2. Ler o header gerado para um arquivo de entrada vazia (relPath = "")
	header := GerarImportHeaderWeb("")
	if header == "" {
		t.Fatal(`GerarImportHeaderWeb("") retornou string vazia`)
	}

	// 3. Extrair os nomes listados no import header e checar duplicatas
	importBlock := strings.TrimSpace(strings.Split(header, "from")[0])
	importBlock = strings.TrimPrefix(importBlock, "import {")
	importBlock = strings.TrimSpace(strings.TrimSuffix(importBlock, "}"))
	namesInHeader := make(map[string]bool)
	seen := make(map[string]int)
	var duplicados []string
	for _, part := range strings.Split(importBlock, ",") {
		name := strings.TrimSpace(part)
		if name != "" {
			seen[name]++
			if seen[name] == 2 {
				duplicados = append(duplicados, name)
			}
			namesInHeader[name] = true
		}
	}

	if len(duplicados) > 0 {
		sort.Strings(duplicados)
		t.Errorf("Identificadores duplicados encontrados em GerarImportHeaderWeb: %v", duplicados)
	}

	// 4. Validar que 'requisitar' está presente no header
	if !namesInHeader["requisitar"] {
		t.Errorf("Símbolo 'requisitar' ausente no header gerado por GerarImportHeaderWeb")
	}

	// 5. Validar coerência: cada nome no header deve existir no runtime-web.js
	var faltantesNoRuntime []string
	for name := range namesInHeader {
		if !actualExports[name] {
			faltantesNoRuntime = append(faltantesNoRuntime, name)
		}
	}
	if len(faltantesNoRuntime) > 0 {
		sort.Strings(faltantesNoRuntime)
		t.Errorf("Os seguintes nomes no import header NÃO foram encontrados como exports no runtime-web.js: %v", faltantesNoRuntime)
	}

	// 6. Validar coerência com FallbackRuntimeWebJS
	fallbackExports := extrairExports(FallbackRuntimeWebJS)
	var faltantesNoFallback []string
	for name := range namesInHeader {
		if !fallbackExports[name] {
			faltantesNoFallback = append(faltantesNoFallback, name)
		}
	}
	if len(faltantesNoFallback) > 0 {
		sort.Strings(faltantesNoFallback)
		t.Errorf("Os seguintes nomes no import header NÃO foram encontrados como exports em FallbackRuntimeWebJS: %v", faltantesNoFallback)
	}

	// 7. Bônus: reportar exports do runtime-web.js que NÃO estão no header (potencial acréscimo esquecido)
	var ausentesNoHeader []string
	for name := range actualExports {
		if !namesInHeader[name] {
			ausentesNoHeader = append(ausentesNoHeader, name)
		}
	}
	if len(ausentesNoHeader) > 0 {
		sort.Strings(ausentesNoHeader)
		t.Logf("AVISO: os seguintes exports do runtime-web.js NÃO estão no import header (verificar se intencional): %v", ausentesNoHeader)
	}
}

// TestRuntimeWebBlindagemECicloDeVida valida que as blindagens de segurança (R2-WEB-01..06)
// e o ciclo de vida de sinais e efeitos (R3-WEB-01..03) estão presentes no runtime-web.js e no FallbackRuntimeWebJS.
func TestRuntimeWebBlindagemECicloDeVida(t *testing.T) {
	runtimeWebPath := filepath.Join("..", "stdlib", "web", "runtime-web.js")
	raw, err := os.ReadFile(runtimeWebPath)
	if err != nil {
		t.Fatalf("erro ao ler runtime-web.js: %v", err)
	}

	arquivos := map[string]string{
		"runtime-web.js":       string(raw),
		"FallbackRuntimeWebJS": FallbackRuntimeWebJS,
	}

	for nome, conteudo := range arquivos {
		t.Run(nome, func(t *testing.T) {
			// 1. armazem usa Object.create(null) e Object.hasOwn(objeto, chave), sem const sinais = {}
			if strings.Contains(conteudo, "const sinais = {}") {
				t.Errorf("[%s] armazem não deve usar 'const sinais = {}'", nome)
			}
			if !strings.Contains(conteudo, "const sinais = Object.create(null)") {
				t.Errorf("[%s] armazem deve usar 'const sinais = Object.create(null)'", nome)
			}
			if !strings.Contains(conteudo, "Object.hasOwn(objeto, chave)") {
				t.Errorf("[%s] armazem deve usar 'Object.hasOwn(objeto, chave)'", nome)
			}

			// 2. urlSegura existe e filtra protocolos
			if !strings.Contains(conteudo, "function urlSegura(") {
				t.Errorf("[%s] helper urlSegura ausente", nome)
			}
			for _, proto := range []string{"http", "https", "mailto", "tel"} {
				if !strings.Contains(conteudo, proto) {
					t.Errorf("[%s] urlSegura deve validar protocolo '%s'", nome, proto)
				}
			}
			if !strings.Contains(conteudo, "urlSegura(") {
				t.Errorf("[%s] chamada de urlSegura ausente", nome)
			}

			// 3. ef.desmontar remove ef de filaEfeitos e marca desmontado = true
			if !strings.Contains(conteudo, "filaEfeitos.delete(ef)") {
				t.Errorf("[%s] ef.desmontar deve remover ef de filaEfeitos ('filaEfeitos.delete(ef)')", nome)
			}
			if !strings.Contains(conteudo, "ef.desmontado = true") {
				t.Errorf("[%s] ef.desmontar deve marcar 'ef.desmontado = true'", nome)
			}
			if !strings.Contains(conteudo, "if (ef.desmontado) return") {
				t.Errorf("[%s] ef.run deve verificar 'if (ef.desmontado) return'", nome)
			}

			// 4. usarConsulta usa AbortController e verifica if (!res.ok)
			if !strings.Contains(conteudo, "new AbortController()") {
				t.Errorf("[%s] usarConsulta deve instanciar 'new AbortController()'", nome)
			}
			if !strings.Contains(conteudo, "!res.ok") {
				t.Errorf("[%s] usarConsulta deve verificar '!res.ok'", nome)
			}

			// 5. registroSinais NÃO adiciona todo sinal globalmente de forma incondicional
			if strings.Contains(conteudo, "\n  registroSinais.add(ler);\n") {
				t.Errorf("[%s] registroSinais não deve reter sinais incondicionalmente em sinal()", nome)
			}
			if !strings.Contains(conteudo, "__HARPIA_DEVTOOLS__.ativo") {
				t.Errorf("[%s] registroSinais deve verificar '__HARPIA_DEVTOOLS__.ativo' antes de registrar", nome)
			}
		})
	}
}
