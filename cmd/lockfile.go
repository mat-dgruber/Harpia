//go:build !js && !wasm

package cmd

import (
	"bufio"
	"bytes"
	"encoding/json"
	"os"
	"strings"
)

// PacoteLockInfo contém as informações detalhadas de versão registrada, URL remota
// de download e o hash de verificação de integridade SHA-256 do pacote travado.
type PacoteLockInfo struct {
	Versao string `json:"versao"`
	URL    string `json:"url"`
	SHA256 string `json:"sha256"`
}

// Lockfile representa a estrutura em memória de um arquivo de lock de dependências
// do Harpia (harpia.lock, harpia.lock.json, pacote.lock, pacote.lock.json).
type Lockfile struct {
	VersaoLock string                    `json:"versao_lock"`
	Pacotes    map[string]PacoteLockInfo `json:"pacotes"`
}

// parseLockfile realiza a desserialização do conteúdo de um arquivo de lockfile.
// Tenta primeiramente efetuar o unmarshal do padrão padrão JSON. Caso o conteúdo não
// seja JSON válido, executa fallback para a análise linha-a-linha declarativa em
// formato chave-valor HRP (ex: versao_lock = "1.0", [nome-do-pacote], etc).
func parseLockfile(conteudo []byte) (*Lockfile, error) {
	var lock Lockfile
	if err := json.Unmarshal(conteudo, &lock); err == nil && (lock.VersaoLock != "" || lock.Pacotes != nil) {
		if lock.Pacotes == nil {
			lock.Pacotes = make(map[string]PacoteLockInfo)
		}
		return &lock, nil
	}

	return parseLockfileHRP(conteudo)
}

// parseLockfileHRP realiza o parse do formato declarativo chave-valor do Harpia (.lock/.hrp).
// O analisador processa comentários (`#` ou `//`), seções delimitadas por colchetes `[pacote]`
// e atribuições no formato `chave = valor`.
func parseLockfileHRP(conteudo []byte) (*Lockfile, error) {
	lock := &Lockfile{
		Pacotes: make(map[string]PacoteLockInfo),
	}

	scanner := bufio.NewScanner(bytes.NewReader(conteudo))
	var pacoteAtual string
	var infoAtual PacoteLockInfo

	salvarPacoteAtual := func() {
		if pacoteAtual != "" {
			lock.Pacotes[pacoteAtual] = infoAtual
		}
	}

	for scanner.Scan() {
		linha := strings.TrimSpace(scanner.Text())

		if linha == "" || strings.HasPrefix(linha, "#") || strings.HasPrefix(linha, "//") {
			continue
		}

		if strings.HasPrefix(linha, "[") && strings.HasSuffix(linha, "]") {
			salvarPacoteAtual()
			pacoteAtual = strings.TrimSpace(linha[1 : len(linha)-1])
			infoAtual = PacoteLockInfo{}
			continue
		}

		partes := strings.SplitN(linha, "=", 2)
		if len(partes) != 2 {
			continue
		}

		chave := strings.TrimSpace(partes[0])
		valor := strings.TrimSpace(partes[1])
		valor = strings.Trim(valor, "\"'")

		if pacoteAtual == "" {
			if chave == "versao_lock" {
				lock.VersaoLock = valor
			}
		} else {
			switch chave {
			case "versao":
				infoAtual.Versao = valor
			case "url":
				infoAtual.URL = valor
			case "sha256":
				infoAtual.SHA256 = valor
			}
		}
	}

	salvarPacoteAtual()

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return lock, nil
}

// salvarLockfile grava a estrutura Lockfile em disco formatada como JSON indentado.
func salvarLockfile(caminho string, lock *Lockfile) error {
	if lock.Pacotes == nil {
		lock.Pacotes = make(map[string]PacoteLockInfo)
	}

	dados, err := json.MarshalIndent(lock, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(caminho, dados, 0644)
}
