const vscode = require('vscode');
const { LanguageClient } = require('vscode-languageclient/node');
const cp = require('child_process');

let client;

function activate(context) {
    // Configuração do executável da CLI para rodar o LSP em segundo plano
    let serverOptions = {
        run: { command: 'harpia', args: ['lsp'] },
        debug: { command: 'harpia', args: ['lsp'] }
    };

    // Opções de comunicação com a IDE
    let clientOptions = {
        documentSelector: [{ scheme: 'file', language: 'harpia' }],
        synchronize: {
            fileEvents: vscode.workspace.createFileSystemWatcher('**/*.hrp')
        }
    };

    // Instancia o cliente LSP
    client = new LanguageClient(
        'harpiaLSP',
        'Harpia Language Server',
        serverOptions,
        clientOptions
    );

    // Inicializa o cliente LSP
    client.start();
    console.log('🐞 Harpia Extension ativa e conectada ao LSP nativo!');

    // ponytail: item discreto na barra de status para acesso rápido ao painel lateral
    const statusBarItem = vscode.window.createStatusBarItem(vscode.StatusBarAlignment.Left, 100);
    statusBarItem.text = '⚡ Harpia';
    statusBarItem.tooltip = 'Abrir Painel de Controle Harpia';
    statusBarItem.command = 'workbench.view.extension.harpia-explorer';
    statusBarItem.show();
    context.subscriptions.push(statusBarItem);

    const obterWorkspaceRoot = () => {
        return vscode.workspace.workspaceFolders && vscode.workspace.workspaceFolders.length > 0
            ? vscode.workspace.workspaceFolders[0].uri.fsPath
            : null;
    };

    const executarScaffolding = (tipo, nome) => {
        const root = obterWorkspaceRoot();
        if (!root) {
            vscode.window.showErrorMessage('Nenhum projeto aberto no espaço de trabalho!');
            return;
        }

        // ponytail: sanitiza entrada para evitar injeção de comandos e aspas para suportar espaços
        const nomeSanitizado = nome.replace(/[^a-zA-Z0-9_\-\s]/g, '');
        if (!nomeSanitizado || nomeSanitizado !== nome) {
            vscode.window.showErrorMessage('Nome inválido! Use apenas letras, números, espaços, hifens ou sublinhados.');
            return;
        }

        const comando = `harpia crie ${tipo} "${nomeSanitizado}"`;
        cp.exec(comando, { cwd: root }, (err, stdout, stderr) => {
            if (err) {
                vscode.window.showErrorMessage(`Erro ao gerar ${tipo}: ${stderr || err.message}`);
                return;
            }
            vscode.window.showInformationMessage(stdout.trim());

            // ponytail: tenta extrair o caminho do arquivo criado da saída e abri-lo automaticamente
            const match = stdout.match(/(?:[a-zA-Z]:)?[\\\w/\-._]+\.hrp/);
            const path = require('path');
            let caminhoArquivo;
            if (match) {
                caminhoArquivo = path.join(root, match[0]);
            } else {
                const pastas = { rota: 'rotas', componente: 'componentes', modelo: 'modelos' };
                const pasta = pastas[tipo] || `${tipo}s`;
                caminhoArquivo = path.join(root, 'src', pasta, `${nomeSanitizado}.hrp`);
            }
            vscode.workspace.openTextDocument(caminhoArquivo).then(
                doc => vscode.window.showTextDocument(doc),
                () => {
                    const pastas = { rota: 'rotas', componente: 'componentes', modelo: 'modelos' };
                    const pasta = pastas[tipo] || `${tipo}s`;
                    const fallbackPath = path.join(root, pasta, `${nomeSanitizado}.hrp`);
                    vscode.workspace.openTextDocument(fallbackPath).then(doc => vscode.window.showTextDocument(doc), () => {});
                }
            );
        });
    };

    context.subscriptions.push(
        vscode.commands.registerCommand('harpia.criarRota', async () => {
            const nome = await vscode.window.showInputBox({ prompt: 'Digite o nome da rota (ex: contato)' });
            if (nome) executarScaffolding('rota', nome);
        })
    );

    context.subscriptions.push(
        vscode.commands.registerCommand('harpia.criarComponente', async () => {
            const nome = await vscode.window.showInputBox({ prompt: 'Digite o nome do componente (ex: Botao)' });
            if (nome) executarScaffolding('componente', nome);
        })
    );

    context.subscriptions.push(
        vscode.commands.registerCommand('harpia.criarModelo', async () => {
            const nome = await vscode.window.showInputBox({ prompt: 'Digite o nome do modelo de domínio (ex: usuario)' });
            if (nome) executarScaffolding('modelo', nome);
        })
    );

    // Registra o provedor do Painel de Controle (Webview)
    const provedorPainel = new HarpiaDashboardProvider(context.extensionUri, obterWorkspaceRoot, executarScaffolding);
    context.subscriptions.push(
        vscode.window.registerWebviewViewProvider('harpia.dashboard', provedorPainel)
    );

    // Formatador que delega para a CLI Harpia (`harpia formatar`).
    // Resolve o problema "Não há formatador para .hrp" ao pressionar Alt+Shift+F
    // e permite format-on-save automaticamente.
    context.subscriptions.push(
        vscode.languages.registerDocumentFormattingEditProvider(
            { scheme: 'file', language: 'harpia' },
            {
                provideDocumentFormattingEdits(document) {
                    return new Promise((resolve) => {
                        const texto = document.getText();
                        cp.execFile('harpia', ['formatar'], { cwd: obterWorkspaceRoot() || undefined, input: texto, maxBuffer: 10 * 1024 * 1024 }, (err, stdout) => {
                            if (err || !stdout) {
                                vscode.window.showWarningMessage('Harpia formatter indisponível. Verifique se a CLI `harpia` está instalada e atualizada.');
                                return resolve([]);
                            }
                            const range = new vscode.Range(
                                document.positionAt(0),
                                document.positionAt(texto.length)
                            );
                            resolve([vscode.TextEdit.replace(range, stdout)]);
                        });
                    });
                }
            }
        )
    );

    // Comando manual "Harpia: Formatar Documento"
    context.subscriptions.push(
        vscode.commands.registerCommand('harpia.formatar', async () => {
            const editor = vscode.window.activeTextEditor;
            if (editor && editor.document.languageId === 'harpia') {
                vscode.commands.executeCommand('editor.action.formatDocument');
            }
        })
    );

    // ponytail: hover documentation completa para palavras-chave da stdlib e definições locais do usuário
    const documentacoesHover = {
        'func': {
            titulo: 'Declaração de Função (`func`)',
            descricao: 'Define um bloco de código reutilizável que pode receber argumentos e retornar um valor.',
            exemplo: '```harpia\nfunc soma(a, b) {\n    retorne a + b;\n}\n```'
        },
        'funcao': {
            titulo: 'Declaração de Função (`funcao`)',
            descricao: 'Define um bloco de código reutilizável com tipagem opcional de parâmetros e retorno.',
            exemplo: '```harpia\nfuncao somar(a: Inteiro, b: Inteiro = 0) -> Inteiro {\n    retorne a + b\n}\n```'
        },
        'retorne': {
            titulo: 'Retorno de Valor (`retorne`)',
            descricao: 'Encerra a execução de uma função e retorna um valor para quem a chamou.',
            exemplo: '```harpia\nretorne resultado;\n```'
        },
        'se': {
            titulo: 'Condicional (`se`)',
            descricao: 'Executa um bloco de código se a condição fornecida for verdadeira (`Verdadeiro`). Parênteses na condição são proibidos.',
            exemplo: '```harpia\nse temperatura > 30 {\n    imprimir("Está quente!");\n}\n```'
        },
        'senao': {
            titulo: 'Condicional Alternativa (`senao`)',
            descricao: 'Executa um bloco de código caso as condições anteriores (`se` ou `senao se`) resultem em `Falso`.',
            exemplo: '```harpia\nse idade >= 18 {\n    imprimir("Maior");\n} senao {\n    imprimir("Menor");\n}\n```'
        },
        'enquanto': {
            titulo: 'Laço de Repetição (`enquanto`)',
            descricao: 'Executa repetidamente um bloco de código enquanto a condição fornecida for verdadeira.',
            exemplo: '```harpia\nenquanto contador < 10 {\n    imprimir(contador);\n    contador = contador + 1;\n}\n```'
        },
        'para': {
            titulo: 'Laço de Repetição (`para ... em`)',
            descricao: 'Itera sobre elementos de uma lista, tupla ou chaves de um mapa.',
            exemplo: '```harpia\npara fruta em frutas {\n    imprimir(fruta)\n}\n```'
        },
        'em': {
            titulo: 'Operador de Iteração (`em`)',
            descricao: 'Especifica a coleção a ser percorrida no laço `para`.',
            exemplo: '```harpia\npara i em [1, 2, 3] {\n    imprimir(i)\n}\n```'
        },
        'var': {
            titulo: 'Declaração de Variável (`var`)',
            descricao: 'Declara uma nova variável mutável no escopo atual, suportando tipagem opcional.',
            exemplo: '```harpia\nvar nome: Texto = "Harpia";\n```'
        },
        'constante': {
            titulo: 'Declaração de Constante (`constante`)',
            descricao: 'Declara uma constante imutável que deve ser obrigatoriamente inicializada no momento da declaração.',
            exemplo: '```harpia\nconstante PI = 3.14159;\n```'
        },
        'imprima': {
            titulo: 'Função de Saída (`imprima`)',
            descricao: 'Função global embutida para exibir valores e mensagens na saída padrão.',
            exemplo: '```harpia\nimprima("Olá, Mundo!", variavel);\n```'
        },
        'imprimir': {
            titulo: 'Função de Saída (`imprimir`)',
            descricao: 'Função global embutida de sistema para exibir valores e mensagens no console.',
            exemplo: '```harpia\nimprimir("Execução concluída.")\n```'
        },
        'Verdadeiro': {
            titulo: 'Literal Booleano (`Verdadeiro`)',
            descricao: 'Representa o valor de verdade lógico positivo (equivalente a `true`).',
            exemplo: '```harpia\nvar ligado = Verdadeiro;\n```'
        },
        'Falso': {
            titulo: 'Literal Booleano (`Falso`)',
            descricao: 'Representa o valor de verdade lógico negativo (equivalente a `false`).',
            exemplo: '```harpia\nvar ligado = Falso;\n```'
        },
        'Nulo': {
            titulo: 'Literal Nulo (`Nulo`)',
            descricao: 'Representa a ausência intencional de qualquer valor de objeto.',
            exemplo: '```harpia\nvar valor: Nulo = Nulo;\n```'
        },
        'classe': {
            titulo: 'Declaração de Classe (`classe`)',
            descricao: 'Define uma nova classe com suporte a atributos, métodos estáticos/de instância e herança.',
            exemplo: '```harpia\nclasse Animal {\n    inicializar(self, nome: Texto) {\n        self.nome = nome\n    }\n}\n```'
        },
        'estende': {
            titulo: 'Herança de Classe (`estende`)',
            descricao: 'Indica a herança de uma classe pai para herança simples de comportamento e propriedades.',
            exemplo: '```harpia\nclasse Cachorro estende Animal {\n    falar(self) -> Texto {\n        retorne "Au!"\n    }\n}\n```'
        },
        'inicializar': {
            titulo: 'Construtor de Classe (`inicializar`)',
            descricao: 'Método construtor padrão executado ao instanciar uma nova classe. Deve receber `self` obrigatoriamente como primeiro parâmetro.',
            exemplo: '```harpia\ninicializar(self, marca: Texto, ano: Inteiro) {\n    self.marca = marca\n    self.ano = ano\n}\n```'
        },
        'self': {
            titulo: 'Referência de Instância (`self`)',
            descricao: 'Palavra-chave que se refere à instância de classe atual nos métodos de objeto.',
            exemplo: '```harpia\nself.nome = "Rex"\n```'
        },
        'estatico': {
            titulo: 'Método Estático (`estatico`)',
            descricao: 'Declara um método pertencente à classe em si, em vez de pertencer às instâncias individuais.',
            exemplo: '```harpia\nestatico buzinar() {\n    imprimir("Bibi!")\n}\n```'
        },
        'tente': {
            titulo: 'Tratamento de Exceção (`tente`)',
            descricao: 'Inicia um bloco de código monitorado para tratamento de erros em runtime.',
            exemplo: '```harpia\ntente {\n    var res = 10 / 0\n} capture (erro) {\n    imprimir(erro.mensagem)\n}\n```'
        },
        'capture': {
            titulo: 'Tratamento de Exceção (`capture`)',
            descricao: 'Captura e trata erros gerados no bloco `tente`. Recebe um objeto de erro com propriedades `mensagem`, `codigoErro` e `sugestao`.',
            exemplo: '```harpia\ncapture (erro) {\n    imprimir("Erro: " + erro.mensagem)\n}\n```'
        },
        'finalmente': {
            titulo: 'Tratamento de Exceção (`finalmente`)',
            descricao: 'Define o bloco final que é garantido de rodar independente do sucesso ou erro no bloco `tente`.',
            exemplo: '```harpia\nfinalmente {\n    imprimir("Concluído.")\n}\n```'
        },
        'assincrono': {
            titulo: 'Corotina Assíncrona (`assincrono`)',
            descricao: 'Declara funções assíncronas que rodam de forma concorrente leve no Event Loop do runtime.',
            exemplo: '```harpia\nassincrono funcao baixar(url) {\n    retorne requisitar("GET", url)\n}\n```'
        },
        'aguarde': {
            titulo: 'Aguardar Operação (`aguarde`)',
            descricao: 'Aguarda de forma não-bloqueante a conclusão de corotinas ou recebimento de dados de canais.',
            exemplo: '```harpia\nvar dados = aguarde requisitar("GET", url)\n```'
        },
        'Canal': {
            titulo: 'Canal Concorrente (`Canal`)',
            descricao: 'Cria um canal FIFO thread-safe para comunicação e sincronização leve entre corotinas (modelo CSP).',
            exemplo: '```harpia\nvar canal = Canal()\naguarde canal.enviar(42)\nvar val = aguarde canal.receber()\n```'
        },
        'testar': {
            titulo: 'Bloco de Testes (`testar`)',
            descricao: 'Declara um cenário de teste unitário integrado nativo na linguagem Harpia.',
            exemplo: '```harpia\ntestar "soma simples" {\n    assegura(somar(1, 1) == 2)\n}\n```'
        },
        'assegura': {
            titulo: 'Asserção de Teste (`assegura`)',
            descricao: 'Verifica se uma condição de teste é verdadeira. Dispara erro de falha de teste se for falsa.',
            exemplo: '```harpia\nassegura(valor == Verdadeiro)\n```'
        },
        'importar': {
            titulo: 'Importação (`importar`)',
            descricao: 'Importa dependências, classes ou bibliotecas padrão para o arquivo atual.',
            exemplo: '```harpia\nimportar { Matematica } de "matematica"\n```'
        },
        'de': {
            titulo: 'Origem de Importação (`de`)',
            descricao: 'Indica de qual módulo ou arquivo local/remoto um recurso está sendo importado.',
            exemplo: '```harpia\nimportar { sinal } de "web"\n```'
        },
        'sinal': {
            titulo: 'Sinal Reativo (`sinal`)',
            descricao: 'Cria e gerencia estados reativos de granularidade fina para atualizações precisas do DOM no frontend. Retorna uma tupla contendo a função de leitura e a função de escrita.',
            exemplo: '```harpia\nvar [contador, definirContador] = sinal(0)\n```'
        },
        'sinalPersistente': {
            titulo: 'Sinal Persistente (`sinalPersistente`)',
            descricao: 'Cria um sinal reativo cujo valor é sincronizado automaticamente e de forma síncrona com o Local Storage.',
            exemplo: '```harpia\nvar [tema, setTema] = sinalPersistente("tema", "claro")\n```'
        },
        'sinalDebounce': {
            titulo: 'Sinal com Debounce (`sinalDebounce`)',
            descricao: 'Cria um sinal reativo cujo valor sofre um atraso de propagação em milissegundos para evitar re-renderizações excessivas.',
            exemplo: '```harpia\nvar [busca, setBusca] = sinalDebounce("", 300)\n```'
        },
        'recurso': {
            titulo: 'Recurso Assíncrono (`recurso`)',
            descricao: 'Gerencia estados de requisições de APIs assíncronas fornecendo propriedades reativas como `.carregando()`, `.erro()` e `.ok()`.',
            exemplo: '```harpia\nvar usuario = recurso(funcao() {\n    retorne requisitar("GET", "https://api.com")\n})\n```'
        },
        'armazem': {
            titulo: 'Armazém de Estado (`armazem`)',
            descricao: 'Cria um contêiner reativo mutável compartilhado para gerenciamento de estado global no frontend.',
            exemplo: '```harpia\nexportar var carrinho = armazem({\n    itens: [],\n    total: 0\n})\n```'
        },
        'efeito': {
            titulo: 'Efeito Reativo (`efeito`)',
            descricao: 'Registra um efeito colateral que é executado automaticamente sempre que qualquer sinal reativo lido dentro dele mudar.',
            exemplo: '```harpia\nefeito(funcao() {\n    imprimir("Mudou para: " + contagem())\n})\n```'
        },
        'derivado': {
            titulo: 'Estado Derivado (`derivado`)',
            descricao: 'Cria um valor reativo somente leitura calculado dinamicamente a partir de outros sinais, recalculado apenas quando necessário.',
            exemplo: '```harpia\nvar dobro = derivado(funcao() => contador() * 2)\n```'
        }
    };

    context.subscriptions.push(
        vscode.languages.registerHoverProvider({ scheme: 'file', language: 'harpia' }, {
            provideHover(document, position) {
                const range = document.getWordRangeAtPosition(position);
                if (!range) return null;

                const palavra = document.getText(range);

                // 1. Tentar documentação nativa baseada no manual
                const docNativa = documentacoesHover[palavra];
                if (docNativa) {
                    const conteudo = new vscode.MarkdownString();
                    conteudo.appendMarkdown(`### **${docNativa.titulo}**\n\n`);
                    conteudo.appendMarkdown(`${docNativa.descricao}\n\n`);
                    conteudo.appendMarkdown(`**Exemplo de uso:**\n${docNativa.exemplo}`);
                    return new vscode.Hover(conteudo);
                }

                // 2. Resolver recursivamente (local, importações e herança profunda)
                const fs = require('fs');
                const path = require('path');
                const root = obterWorkspaceRoot() || '';
                const resultado = resolverPalavraEstendida(palavra, document.uri.fsPath, root);

                if (resultado) {
                    const conteudo = new vscode.MarkdownString();
                    const eLocal = resultado.caminho === document.uri.fsPath;
                    const arquivoRelativo = path.relative(root, resultado.caminho);
                    const origemTexto = eLocal ? 'Local' : `Importada de \`${arquivoRelativo}\``;

                    let cabecalho = `### **${palavra}** (${resultado.tipo} — ${origemTexto})\n\n`;
                    conteudo.appendMarkdown(cabecalho);

                    if (resultado.cadeiaHeranca && resultado.cadeiaHeranca.length > 1) {
                        conteudo.appendMarkdown(`**Hierarquia de Herança:**\n`);
                        conteudo.appendMarkdown(`\`${resultado.cadeiaHeranca.join(' ➔ ')}\`\n\n`);
                    }

                    conteudo.appendMarkdown(`\`\`\`harpia\n${resultado.assinatura}\n\`\`\`\n\n`);

                    if (resultado.comentarios) {
                        conteudo.appendMarkdown(`**Documentação:**\n${resultado.comentarios}`);
                    }
                    return new vscode.Hover(conteudo);
                }

                return null;
            }
        })
    );

    // ponytail: F12 (Go to Definition) para pular para declarações locais ou importadas
    context.subscriptions.push(
        vscode.languages.registerDefinitionProvider({ scheme: 'file', language: 'harpia' }, {
            provideDefinition(document, position) {
                const range = document.getWordRangeAtPosition(position);
                if (!range) return null;

                const palavra = document.getText(range);
                const root = obterWorkspaceRoot() || '';

                if (documentacoesHover[palavra]) return null;

                const resultado = resolverPalavraEstendida(palavra, document.uri.fsPath, root);
                if (resultado && resultado.caminho) {
                    const fs = require('fs');
                    if (fs.existsSync(resultado.caminho)) {
                        const texto = fs.readFileSync(resultado.caminho, 'utf-8');
                        const linhas = texto.split('\n');

                        const regexFunc = new RegExp(`(?:funcao|func)\\s+\\b${palavra}\\b`);
                        const regexClasse = new RegExp(`classe\\s+\\b${palavra}\\b`);
                        const regexVar = new RegExp(`(?:var|constante)\\s+\\b${palavra}\\b`);

                        for (let i = 0; i < linhas.length; i++) {
                            const linha = linhas[i];
                            if (regexFunc.test(linha) || regexClasse.test(linha) || regexVar.test(linha)) {
                                return new vscode.Location(
                                    vscode.Uri.file(resultado.caminho),
                                    new vscode.Position(i, 0)
                                );
                            }
                        }
                    }
                    return new vscode.Location(
                        vscode.Uri.file(resultado.caminho),
                        new vscode.Position(0, 0)
                    );
                }
                return null;
            }
        })
    );

    // ponytail: CodeLenses flutuantes para executar cenários de testes isolados de forma rápida
    context.subscriptions.push(
        vscode.languages.registerCodeLensProvider({ scheme: 'file', language: 'harpia' }, {
            provideCodeLenses(document) {
                const codeLenses = [];
                const texto = document.getText();
                const regexTeste = /testar\s+["']([^"']+)["']\s*\{/g;
                let match;

                while ((match = regexTeste.exec(texto)) !== null) {
                    const nomeTeste = match[1];
                    const posicao = document.positionAt(match.index);
                    const range = new vscode.Range(posicao, posicao);

                    const lens = new vscode.CodeLens(range, {
                        title: `▶️ Executar teste "${nomeTeste}"`,
                        command: 'harpia.rodarTeste',
                        arguments: [document.uri.fsPath, nomeTeste]
                    });
                    codeLenses.push(lens);
                }
                return codeLenses;
            }
        })
    );

    context.subscriptions.push(
        vscode.commands.registerCommand('harpia.rodarTeste', (caminhoArquivo, nomeTeste) => {
            let term = vscode.window.terminals.find(t => t.name === 'Harpia Testes') || vscode.window.createTerminal('Harpia Testes');
            term.show();
            term.sendText(`harpia testar "${caminhoArquivo}" --filtro="${nomeTeste}"`);
        })
    );

    // ponytail: autocomplete (CompletionItemProvider) inteligente para imports e módulos nativos do Harpia
    const modulosNativos = {
        'web': ['sinal', 'sinalPersistente', 'sinalDebounce', 'efeito', 'derivado', 'recurso', 'armazem', 'Provedor', 'injetar', 'FronteiraDeErro', 'ListaVirtual', 'GradeDeDados', 'preguicoso', 'Suspense'],
        'resiliencia': ['novo_disjuntor', 'novo_limite_de_taxa', 'nova_retentativa'],
        'telemetria': ['novo_tracer', 'nova_metrica'],
        'bd': ['conectarQdrant', 'tabela'],
        'ia': ['validar_resposta', 'criar_agente', 'gerar_texto']
    };

    context.subscriptions.push(
        vscode.languages.registerCompletionItemProvider({ scheme: 'file', language: 'harpia' }, {
            provideCompletionItems(document, position) {
                const rangeLinha = new vscode.Range(new vscode.Position(position.line, 0), position);
                const textoLinha = document.getText(rangeLinha);

                // 1. Sugerir nomes dos módulos nativos quando o usuário digita 'de "' ou 'de ''
                const matchDe = textoLinha.match(/de\s+["']([^"']*)$/);
                if (matchDe) {
                    return Object.keys(modulosNativos).map(nomeModulo => {
                        const item = new vscode.CompletionItem(nomeModulo, vscode.CompletionItemKind.Module);
                        item.detail = `Módulo Nativo do Harpia`;
                        return item;
                    });
                }

                // 2. Sugerir funções de um módulo nativo quando o usuário estiver dentro de 'importar { ... } de "modulo"'
                const matchImport = document.lineAt(position.line).text.match(/importar\s*\{([^}]*)\}\s*de\s*["']([^"']+)["']/);
                if (matchImport) {
                    const nomeModulo = matchImport[2];
                    const funcoesModulo = modulosNativos[nomeModulo];
                    if (funcoesModulo) {
                        return funcoesModulo.map(func => {
                            const item = new vscode.CompletionItem(func, vscode.CompletionItemKind.Function);
                            item.detail = `Função do módulo nativo "${nomeModulo}"`;
                            return item;
                        });
                    }
                }

                // 3. Sugerir funções/palavras globais em qualquer parte do código (fallback)
                const itensGlobais = [
                    ...Object.keys(documentacoesHover).map(palavra => {
                        const item = new vscode.CompletionItem(palavra, vscode.CompletionItemKind.Keyword);
                        item.detail = `Palavra-chave do Harpia`;
                        return item;
                    }),
                    new vscode.CompletionItem('imprimir', vscode.CompletionItemKind.Function),
                    new vscode.CompletionItem('imprima', vscode.CompletionItemKind.Function),
                    new vscode.CompletionItem('assegura', vscode.CompletionItemKind.Function)
                ];
                return itensGlobais;
            }
        }, '"', "'", '{', ',')
    );

    // ponytail: RenameProvider local que permite renomear variáveis e funções no arquivo com segurança usando Regex
    context.subscriptions.push(
        vscode.languages.registerRenameProvider({ scheme: 'file', language: 'harpia' }, {
            provideRenameEdits(document, position, newName) {
                const range = document.getWordRangeAtPosition(position);
                if (!range) return null;

                const oldName = document.getText(range);

                // Evita renomear palavras-chave nativas do framework
                if (documentacoesHover[oldName] || oldName === 'imprimir' || oldName === 'imprima' || oldName === 'assegura') {
                    throw new Error('Não é permitido renomear palavras-chave nativas do Harpia.');
                }

                const workspaceEdit = new vscode.WorkspaceEdit();
                const textoCompleto = document.getText();

                // Regex para encontrar todas as ocorrências isoladas do nome da palavra no arquivo atual
                const regexWord = new RegExp(`\\b${oldName}\\b`, 'g');
                let match;

                while ((match = regexWord.exec(textoCompleto)) !== null) {
                    const localInicio = document.positionAt(match.index);
                    const localFim = document.positionAt(match.index + oldName.length);
                    const rangeSubstituicao = new vscode.Range(localInicio, localFim);

                    workspaceEdit.replace(document.uri, rangeSubstituicao, newName);
                }

                return workspaceEdit;
            }
        })
    );

    // ponytail: Signature Help (Ctrl+Shift+Space) para exibir assinaturas de funções de forma inteligente
    const assinaturasNativas = {
        'sinal': { assinatura: 'sinal(valorInicial)', params: ['valorInicial'], doc: 'Cria um sinal reativo contendo o valor inicial.' },
        'sinalPersistente': { assinatura: 'sinalPersistente(chave: Texto, valorInicial)', params: ['chave', 'valorInicial'], doc: 'Cria um sinal reativo persistido automaticamente no Local Storage.' },
        'sinalDebounce': { assinatura: 'sinalDebounce(valorInicial, tempoEmMs: Inteiro)', params: ['valorInicial', 'tempoEmMs'], doc: 'Cria um sinal reativo com atraso de propagação (debounce).' },
        'efeito': { assinatura: 'efeito(funcao)', params: ['funcao'], doc: 'Cria um efeito colateral que reage a mudanças nos sinais lidos dentro dele.' },
        'derivado': { signature: 'derivado(funcao)', params: ['funcao'], doc: 'Cria um valor reativo derivado (calculado) somente leitura.' },
        'recurso': { assinatura: 'recurso(funcaoAsync)', params: ['funcaoAsync'], doc: 'Cria um sinal assíncrono para consumo de APIs.' },
        'armazem': { assinatura: 'armazem(objeto)', params: ['objeto'], doc: 'Cria um contêiner reativo mutável compartilhado para estado global.' },
        'imprimir': { assinatura: 'imprimir(...valores)', params: ['...valores'], doc: 'Exibe valores de qualquer tipo na saída padrão de depuração.' },
        'imprima': { assinatura: 'imprima(...valores)', params: ['...valores'], doc: 'Exibe valores de qualquer tipo na saída padrão.' },
        'assegura': { assinatura: 'assegura(condicao: Booleano)', params: ['condicao'], doc: 'Asserção de testes unitários que valida se uma condição é verdadeira.' },
        'novo_disjuntor': { assinatura: 'novo_disjuntor(maxFalhas: Inteiro, tempoEsperaMs: Inteiro)', params: ['maxFalhas', 'tempoEsperaMs'], doc: 'Cria um disjuntor (circuit breaker) resiliente.' },
        'novo_limite_de_taxa': { assinatura: 'novo_limite_de_taxa(maxRequisicoes: Inteiro, intervaloMs: Inteiro)', params: ['maxRequisicoes', 'intervaloMs'], doc: 'Cria um limitador de taxa (rate limiter).' },
        'nova_retentativa': { assinatura: 'nova_retentativa(maxTentativas: Inteiro, atrasoMs: Inteiro)', params: ['maxTentativas', 'atrasoMs'], doc: 'Cria um mecanismo de retentativas automáticas.' }
    };

    context.subscriptions.push(
        vscode.languages.registerSignatureHelpProvider({ scheme: 'file', language: 'harpia' }, {
            provideSignatureHelp(document, position) {
                const textoAtePosicao = document.getText(new vscode.Range(new vscode.Position(position.line, 0), position));

                // Encontra a chamada da função anterior na mesma linha
                const matchFuncao = textoAtePosicao.match(/([a-zA-Z_][a-zA-Z0-9_]*)\s*\(([^)]*)$/);
                if (!matchFuncao) return null;

                const nomeFuncao = matchFuncao[1];
                const argumentosDigitados = matchFuncao[2];

                // Descobre em qual parâmetro o usuário está baseado nas vírgulas
                const parametroAtivoIndex = (argumentosDigitados.match(/,/g) || []).length;

                let infoAssinatura;

                // 1. Verifica se é nativa
                if (assinaturasNativas[nomeFuncao]) {
                    infoAssinatura = assinaturasNativas[nomeFuncao];
                } else {
                    // 2. Busca definição local ou importada do usuário
                    const root = obterWorkspaceRoot() || '';
                    const resultado = resolverPalavraEstendida(nomeFuncao, document.uri.fsPath, root);
                    if (resultado && resultado.tipo === 'Função') {
                        const matchParams = resultado.assinatura.match(/\((.*?)\)/);
                        const paramsTexto = matchParams ? matchParams[1] : '';
                        const params = paramsTexto ? paramsTexto.split(',').map(p => p.trim()) : [];
                        infoAssinatura = {
                            assinatura: resultado.assinatura,
                            params,
                            doc: resultado.comentarios || 'Função customizada do usuário.'
                        };
                    }
                }

                if (infoAssinatura) {
                    const signatureHelp = new vscode.SignatureHelp();
                    const sigInfo = new vscode.SignatureInformation(infoAssinatura.assinatura, infoAssinatura.doc);

                    sigInfo.parameters = infoAssinatura.params.map(p => new vscode.ParameterInformation(p));
                    signatureHelp.signatures = [sigInfo];
                    signatureHelp.activeSignature = 0;
                    signatureHelp.activeParameter = Math.min(parametroAtivoIndex, infoAssinatura.params.length - 1);

                    return signatureHelp;
                }
                return null;
            }
        }, '(', ',')
    );

    // ponytail: DocumentColorProvider para exibir caixas de cores e color picker para hexadecimais, RGB, HSL e cores CSS nomeadas nos arquivos .hrp
    context.subscriptions.push(
        vscode.languages.registerColorProvider({ scheme: 'file', language: 'harpia' }, {
            provideDocumentColors(document) {
                const colors = [];
                const texto = document.getText();

                // 1. Cores Hexadecimais (#RRGGBB / #RGB)
                const regexHex = /#([0-9a-fA-F]{3,4}|[0-9a-fA-F]{6}|[0-9a-fA-F]{8})\b/g;
                let match;
                while ((match = regexHex.exec(texto)) !== null) {
                    const hex = match[1];
                    const range = new vscode.Range(document.positionAt(match.index), document.positionAt(match.index + match[0].length));
                    let r = 0, g = 0, b = 0, a = 1;
                    if (hex.length === 3 || hex.length === 4) {
                        r = parseInt(hex[0] + hex[0], 16) / 255;
                        g = parseInt(hex[1] + hex[1], 16) / 255;
                        b = parseInt(hex[2] + hex[2], 16) / 255;
                        if (hex.length === 4) {
                            a = parseInt(hex[3] + hex[3], 16) / 255;
                        }
                    } else if (hex.length === 6 || hex.length === 8) {
                        r = parseInt(hex.substring(0, 2), 16) / 255;
                        g = parseInt(hex.substring(2, 4), 16) / 255;
                        b = parseInt(hex.substring(4, 6), 16) / 255;
                        if (hex.length === 8) {
                            a = parseInt(hex.substring(6, 8), 16) / 255;
                        }
                    }
                    colors.push(new vscode.ColorInformation(range, new vscode.Color(r, g, b, a)));
                }

                // 2. Cores RGB / RGBA
                const regexRgb = /rgba?\(\s*(\d+)\s*,\s*(\d+)\s*,\s*(\d+)\s*(?:,\s*([\d.]+)\s*)?\)/g;
                while ((match = regexRgb.exec(texto)) !== null) {
                    const r = parseInt(match[1]) / 255;
                    const g = parseInt(match[2]) / 255;
                    const b = parseInt(match[3]) / 255;
                    const a = match[4] ? parseFloat(match[4]) : 1.0;
                    const range = new vscode.Range(document.positionAt(match.index), document.positionAt(match.index + match[0].length));
                    colors.push(new vscode.ColorInformation(range, new vscode.Color(r, g, b, a)));
                }

                // 3. Cores HSL / HSLA
                const regexHsl = /hsla?\(\s*(\d+)\s*,\s*([\d.]+)%\s*,\s*([\d.]+)%\s*(?:,\s*([\d.]+)\s*)?\)/g;
                while ((match = regexHsl.exec(texto)) !== null) {
                    const h = parseInt(match[1]);
                    const s = parseFloat(match[2]) / 100;
                    const l = parseFloat(match[3]) / 100;
                    const a = match[4] ? parseFloat(match[4]) : 1.0;

                    const [r, g, b] = hslParaRgb(h, s, l);
                    const range = new vscode.Range(document.positionAt(match.index), document.positionAt(match.index + match[0].length));
                    colors.push(new vscode.ColorInformation(range, new vscode.Color(r, g, b, a)));
                }

                // 4. Cores Nomeadas CSS
                const coresNomeadas = {
                    "red": [255,0,0], "blue": [0,0,255], "green": [0,128,0], "yellow": [255,255,0],
                    "black": [0,0,0], "white": [255,255,255], "gray": [128,128,128], "purple": [128,0,128],
                    "orange": [255,165,0], "pink": [255,192,203], "brown": [165,42,42], "gold": [255,215,0],
                    "silver": [192,192,192], "cyan": [0,255,255], "magenta": [255,0,255], "lime": [0,255,0],
                    "navy": [0,0,128], "teal": [0,128,128], "olive": [128,128,0]
                };

                const nomesOrPattern = Object.keys(coresNomeadas).join('|');
                const regexNomeada = new RegExp(`["'](${nomesOrPattern})["']`, 'g');
                while ((match = regexNomeada.exec(texto)) !== null) {
                    const nome = match[1];
                    const rgb = coresNomeadas[nome];
                    const posicaoInicio = document.positionAt(match.index + 1);
                    const posicaoFim = document.positionAt(match.index + 1 + nome.length);
                    const range = new vscode.Range(posicaoInicio, posicaoFim);
                    colors.push(new vscode.ColorInformation(range, new vscode.Color(rgb[0]/255, rgb[1]/255, rgb[2]/255, 1)));
                }

                return colors;
            },

            provideColorPresentations(color) {
                const r255 = Math.round(color.red * 255);
                const g255 = Math.round(color.green * 255);
                const b255 = Math.round(color.blue * 255);
                const a = color.alpha;

                // Apresentação Hexadecimal
                const rHex = r255.toString(16).padStart(2, '0');
                const gHex = g255.toString(16).padStart(2, '0');
                const bHex = b255.toString(16).padStart(2, '0');
                const aHex = Math.round(a * 255).toString(16).padStart(2, '0');
                let hex = `#${rHex}${gHex}${bHex}`;
                if (aHex !== 'ff') {
                    hex += aHex;
                }

                // Apresentação RGBA
                const rgba = a === 1
                    ? `rgb(${r255}, ${g255}, ${b255})`
                    : `rgba(${r255}, ${g255}, ${b255}, ${a.toFixed(2)})`;

                // Apresentação HSLA
                const [h, s, l] = rgbParaHsl(color.red, color.green, color.blue);
                const hsla = a === 1
                    ? `hsl(${h}, ${s}%, ${l}%)`
                    : `hsla(${h}, ${s}%, ${l}%, ${a.toFixed(2)})`;

                return [
                    new vscode.ColorPresentation(hex),
                    new vscode.ColorPresentation(rgba),
                    new vscode.ColorPresentation(hsla)
                ];
            }
        })
    );
}

// ponytail: converte HSL para RGB normatizado
function hslParaRgb(h, s, l) {
    let r, g, b;
    if (s === 0) {
        r = g = b = l;
    } else {
        const hue2rgb = (p, q, t) => {
            if (t < 0) t += 1;
            if (t > 1) t -= 1;
            if (t < 1/6) return p + (q - p) * 6 * t;
            if (t < 1/2) return q;
            if (t < 2/3) return p + (q - p) * (2/3 - t) * 6;
            return p;
        };
        const q = l < 0.5 ? l * (1 + s) : l + s - l * s;
        const p = 2 * l - q;
        r = hue2rgb(p, q, h / 360 + 1/3);
        g = hue2rgb(p, q, h / 360);
        b = hue2rgb(p, q, h / 360 - 1/3);
    }
    return [r, g, b];
}

// ponytail: converte RGB normatizado para HSL
function rgbParaHsl(r, g, b) {
    const max = Math.max(r, g, b), min = Math.min(r, g, b);
    let h, s, l = (max + min) / 2;

    if (max === min) {
        h = s = 0;
    } else {
        const d = max - min;
        s = l > 0.5 ? d / (2 - max - min) : d / (max + min);
        switch (max) {
            case r: h = (g - b) / d + (g < b ? 6 : 0); break;
            case g: h = (b - r) / d + 2; break;
            case b: h = (r - g) / d + 4; break;
        }
        h /= 6;
    }

    return [
        Math.round(h * 360),
        Math.round(s * 100),
        Math.round(l * 100)
    ];
}

// ponytail: extrai comentários imediatamente superiores a uma linha para usá-los como javadoc/documentação
function obterComentariosSuperiores(linhas, indiceLinha) {
    const comentarios = [];
    let i = indiceLinha - 1;
    while (i >= 0) {
        const linha = linhas[i].trim();
        if (linha.startsWith('#') || linha.startsWith('//')) {
            const textoComentario = linha.replace(/^(?:#|\/\/\/|\/\/)\s*/, '');
            comentarios.unshift(textoComentario);
        } else if (linha === '') {
            // pula linhas vazias intermediárias
        } else {
            break;
        }
        i--;
    }
    return comentarios.join('\n');
}

// ponytail: tenta resolver uma palavra (classe, função, variável) e sua cadeia de herança/importação
function resolverPalavraEstendida(palavra, caminhoArquivo, root, arquivosProcessados = new Set()) {
    const fs = require('fs');
    const path = require('path');

    if (arquivosProcessados.has(caminhoArquivo)) return null;
    arquivosProcessados.add(caminhoArquivo);

    if (!fs.existsSync(caminhoArquivo)) return null;
    const texto = fs.readFileSync(caminhoArquivo, 'utf-8');
    const linhas = texto.split('\n');

    const regexFunc = new RegExp(`(?:funcao|func)\\s+(${palavra})\\s*\\((.*?)\\)`);
    const regexClasse = new RegExp(`classe\\s+(${palavra})(?:\\s+estende\\s+(\\w+))?\\s*\\{`);
    const regexVar = new RegExp(`(?:var|constante)\\s+(${palavra})\\s*(?::\\s*[\\w<>]+)?`);

    for (let i = 0; i < linhas.length; i++) {
        const linha = linhas[i];

        // 1. É função?
        let match = linha.match(regexFunc);
        if (match) {
            const params = match[2];
            const assinatura = `funcao ${palavra}(${params})`;
            const comentarios = obterComentariosSuperiores(linhas, i);
            return {
                tipo: 'Função',
                assinatura,
                comentarios,
                caminho: caminhoArquivo
            };
        }

        // 2. É classe?
        match = linha.match(regexClasse);
        if (match) {
            const pai = match[2];
            let assinatura = linha.trim().replace(/\{/g, '').trim();
            const comentarios = obterComentariosSuperiores(linhas, i);

            let cadeiaHeranca = [palavra];
            if (pai) {
                // Tenta resolver a classe pai recursivamente na mesma árvore de arquivos
                let classePaiRef = resolverPalavraEstendida(pai, caminhoArquivo, root, new Set(arquivosProcessados));
                if (!classePaiRef) {
                    // Se não achou no mesmo arquivo, tenta ver se o pai foi importado nele
                    const caminhoPaiImportado = buscarCaminhoImportado(pai, texto, caminhoArquivo, root);
                    if (caminhoPaiImportado) {
                        classePaiRef = resolverPalavraEstendida(pai, caminhoPaiImportado, root, new Set(arquivosProcessados));
                    }
                }

                if (classePaiRef && classePaiRef.cadeiaHeranca) {
                    cadeiaHeranca = cadeiaHeranca.concat(classePaiRef.cadeiaHeranca);
                } else if (pai) {
                    cadeiaHeranca.push(pai);
                }
            }

            return {
                tipo: 'Classe',
                assinatura,
                comentarios,
                caminho: caminhoArquivo,
                cadeiaHeranca
            };
        }

        // 3. É variável/constante?
        match = linha.match(regexVar);
        if (match) {
            const assinatura = linha.trim().replace(';', '').trim();
            const comentarios = obterComentariosSuperiores(linhas, i);
            return {
                tipo: 'Variável/Constante',
                assinatura,
                comentarios,
                caminho: caminhoArquivo
            };
        }
    }

    // Se não achou a definição direta, verifica se ela foi importada neste arquivo
    const caminhoImportado = buscarCaminhoImportado(palavra, texto, caminhoArquivo, root);
    if (caminhoImportado) {
        return resolverPalavraEstendida(palavra, caminhoImportado, root, arquivosProcessados);
    }

    return null;
}

// ponytail: busca o caminho absoluto de um arquivo importado de forma relativa ou pela raiz do workspace
function buscarCaminhoImportado(palavra, texto, caminhoArquivoAtual, root) {
    const path = require('path');
    const regexImport = new RegExp(`importar\\s*\\{[^\\}]*\\b${palavra}\\b[^\\}]*\\}\\s*de\\s*["']([^"']+)["']`);
    const match = texto.match(regexImport);
    if (!match) return null;

    const origem = match[1];
    let caminhoResolvido;

    if (origem.startsWith('.')) {
        caminhoResolvido = path.resolve(path.dirname(caminhoArquivoAtual), origem);
    } else {
        caminhoResolvido = path.resolve(root, origem);
    }

    if (!caminhoResolvido.endsWith('.hrp')) {
        caminhoResolvido += '.hrp';
    }
    return caminhoResolvido;
}

class HarpiaDashboardProvider {
    constructor(extensionUri, obterWorkspaceRoot, executarScaffolding) {
        this._extensionUri = extensionUri;
        this.obterWorkspaceRoot = obterWorkspaceRoot;
        this.executarScaffolding = executarScaffolding;
    }

    resolveWebviewView(webviewView, context, token) {
        webviewView.webview.options = {
            enableScripts: true,
            localResourceRoots: [this._extensionUri]
        };

        webviewView.webview.html = this._obterHtmlParaWebview(webviewView.webview);

        // Lida com cliques e interações vindo do painel visual HTML
        webviewView.webview.onDidReceiveMessage(async (data) => {
            const root = this.obterWorkspaceRoot();
            if (!root) {
                vscode.window.showErrorMessage('Abra um projeto Harpia primeiro!');
                return;
            }

            switch (data.command) {
                case 'servir':
                    // Inicia terminal integrado rodando dev server
                    let termServir = vscode.window.terminals.find(t => t.name === 'Harpia Servidor') || vscode.window.createTerminal('Harpia Servidor');
                    termServir.show();
                    termServir.sendText('harpia servir');
                    break;
                case 'empacotar':
                    const entrada = await vscode.window.showInputBox({ prompt: 'Arquivo de entrada (.hrp)', defaultValue: 'main.hrp' });
                    if (!entrada) return;
                    const saida = await vscode.window.showInputBox({ prompt: 'Nome do executável gerado', defaultValue: 'app_compilado' });
                    if (!saida) return;

                    let termEmp = vscode.window.terminals.find(t => t.name === 'Harpia Empacotar') || vscode.window.createTerminal('Harpia Empacotar');
                    termEmp.show();
                    termEmp.sendText(`harpia empacotar --entrada=${entrada} --saida=${saida}`);
                    break;
                case 'stressar':
                    const arqStress = await vscode.window.showInputBox({ prompt: 'Arquivo de teste (.hrp)', defaultValue: 'main.hrp' });
                    if (!arqStress) return;
                    let termStress = vscode.window.terminals.find(t => t.name === 'Harpia Stressar') || vscode.window.createTerminal('Harpia Stressar');
                    termStress.show();
                    termStress.sendText(`harpia stressar --arquivo=${arqStress} -c 10 -r 100`);
                    break;
                case 'scaffold':
                    this.executarScaffolding(data.tipo, data.nome);
                    break;
            }
        });
    }

    _obterHtmlParaWebview(webview) {
        return `<!DOCTYPE html>
        <html lang="pt-br">
        <head>
            <meta charset="UTF-8">
            <meta name="viewport" content="width=device-width, initial-scale=1.0">
            <style>
                body {
                    font-family: var(--vscode-font-family);
                    color: var(--vscode-foreground);
                    padding: 10px;
                }
                .btn {
                    display: block;
                    width: 100%;
                    background: var(--vscode-button-background);
                    color: var(--vscode-button-foreground);
                    border: none;
                    padding: 8px;
                    margin-bottom: 8px;
                    cursor: pointer;
                    text-align: center;
                    border-radius: 4px;
                    font-weight: bold;
                }
                .btn:hover {
                    background: var(--vscode-button-hoverBackground);
                }
                .section-title {
                    font-size: 1.1em;
                    font-weight: bold;
                    margin-top: 15px;
                    margin-bottom: 8px;
                    border-bottom: 1px solid var(--vscode-panel-border);
                    padding-bottom: 4px;
                }
                input {
                    width: 90%;
                    padding: 6px;
                    margin-bottom: 8px;
                    background: var(--vscode-input-background);
                    color: var(--vscode-input-foreground);
                    border: 1px solid var(--vscode-input-border);
                    border-radius: 4px;
                }
            </style>
        </head>
        <body>
            <div class="section-title">Serviços e Execução</div>
            <button class="btn" onclick="enviar('servir')">🚀 Iniciar Servidor Dev</button>

            <div class="section-title">Ferramentas de DevOps</div>
            <button class="btn" onclick="enviar('empacotar')">📦 Empacotar Binário Nativo</button>
            <button class="btn" onclick="enviar('stressar')">🔥 Executar Teste de Estresse</button>

            <div class="section-title">Gerador de Código (Clean Arch)</div>
            <input type="text" id="scaffold-nome" placeholder="Nome do recurso (ex: contato)">
            <button class="btn" onclick="scaffold('rota')">➕ Criar Rota SPA</button>
            <button class="btn" onclick="scaffold('componente')">🧩 Criar Componente UI</button>
            <button class="btn" onclick="scaffold('modelo')">🗂️ Criar Modelo de Domínio</button>

            <script>
                const vscode = acquireVsCodeApi();
                function enviar(cmd) {
                    vscode.postMessage({ command: cmd });
                }
                function scaffold(tipo) {
                    const nome = document.getElementById('scaffold-nome').value.trim();
                    if (!nome) return;
                    vscode.postMessage({ command: 'scaffold', tipo: tipo, nome: nome });
                    document.getElementById('scaffold-nome').value = '';
                }
            </script>
        </body>
        </html>`;
    }
}

function deactivate() {
    if (!client) {
        return undefined;
    }
    return client.stop();
}

module.exports = {
    activate,
    deactivate
};
