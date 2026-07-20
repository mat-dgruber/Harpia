<#
    Script de Instalação do Harpia para Windows
#>
$ErrorActionPreference = "Stop"

function Aplicar-Estilo {
    param (
        [string]$Cor,
        [bool]$Negrito,
        [string]$Texto
    )
    
    $style = ""
    if ($Negrito) { $style += "1;" }
    
    switch ($Cor) {
        "vermelho" { $style += "31" }
        "verde"    { $style += "32" }
        "amarelo"  { $style += "33" }
        "branco"   { $style += "37" }
        default    { $style += "0" }
    }
    
    # Renderiza usando sequências de escape ANSI (suportado no Windows 10/11)
    Write-Host ("`e[" + $style + "m" + $Texto + "`e[0m") -NoNewline
}

function Log-Mensagem {
    param (
        [string]$Nivel,
        [string]$Mensagem
    )
    
    $cor = "branco"
    $negrito = $true
    
    switch ($Nivel) {
        "ERRO"    { $cor = "vermelho" }
        "INFO"    { $cor = "branco" }
        "AVISO"   { $cor = "amarelo" }
        "SUCESSO" { $cor = "verde" }
        default   { $cor = "branco"; $negrito = $false }
    }
    
    Write-Host "[ " -NoNewline
    Aplicar-Estilo -Cor $cor -Negrito $negrito -Texto $Nivel
    Write-Host " ]: " -NoNewline
    Aplicar-Estilo -Cor $cor -Negrito $false -Texto $Mensagem
    Write-Host ""
    
    if ($Nivel -eq "ERRO") {
        Exit 1
    }
}

# Validação de argumentos
if ($args.Count -gt 1) {
    Log-Mensagem "ERRO" "Foi recebido mais argumentos do que o esperado. Caso deseje instalar uma versão específica, use por exemplo: v0.1.0."
}

# Definição de Target (Windows x86_64)
$target = "Windows_x86_64"
$sufixo = ".zip"

$GITHUB = if ($env:GITHUB) { $env:GITHUB } else { "https://github.com" }
$repo_github = "$GITHUB/mat-dgruber/Harpia"
$arquivo_compactado = "$target$sufixo"

if ($args.Count -eq 0) {
    $harpia_uri = "$repo_github/releases/latest/download/$arquivo_compactado"
} else {
    $harpia_uri = "$repo_github/releases/download/$($args[0])/$arquivo_compactado"
}

$raiz_harpia = if ($env:RAIZ_HARPIA) { $env:RAIZ_HARPIA } else { "$env:USERPROFILE\.harpia" }
$diretorio_binario = "$raiz_harpia\bin"
$executavel_zip = "$env:TEMP\harpia.zip"

if (-not (Test-Path $diretorio_binario)) {
    New-Item -ItemType Directory -Force -Path $diretorio_binario | Out-Null
}

# Download
Log-Mensagem "DEBUG" "Iniciando download do arquivo compactado..."
try {
    [Net.ServicePointManager]::SecurityProtocol = [Net.SecurityProtocolType]::Tls12
    Invoke-WebRequest -Uri $harpia_uri -OutFile $executavel_zip -UserAgent "Mozilla/5.0"
} catch {
    if (Test-Path $executavel_zip) { Remove-Item $executavel_zip -Force }
    Log-Mensagem "ERRO" "Falha ao baixar o Harpia de `"$harpia_uri`""
}

# Descompactação
Log-Mensagem "DEBUG" "Iniciando descompactação..."
try {
    Expand-Archive -Path $executavel_zip -DestinationPath $diretorio_binario -Force
    Remove-Item $executavel_zip -Force
} catch {
    if (Test-Path $executavel_zip) { Remove-Item $executavel_zip -Force }
    Log-Mensagem "ERRO" "Falha ao descompactar o arquivo."
}
Log-Mensagem "SUCESSO" "Parece que a descompactação foi um sucesso"

Log-Mensagem "SUCESSO" "Parabéns, agora você tem o Harpia disponível em $diretorio_binario\harpia.exe"

# Atualização do PATH no Windows (Escopo do Usuário)
$pathAtual = [Environment]::GetEnvironmentVariable("PATH", "User")
if ($pathAtual -notlike "*$diretorio_binario*") {
    [Environment]::SetEnvironmentVariable("PATH", "$pathAtual;$diretorio_binario", "User")
    Log-Mensagem "INFO" "Adicionado o caminho `"$diretorio_binario`" ao PATH do Usuário."
    Log-Mensagem "AVISO" "Por favor, reinicie seu terminal para aplicar as alterações do PATH."
} else {
    Log-Mensagem "INFO" "O caminho já está configurado no seu PATH."
}

Write-Host ""
Log-Mensagem "INFO" "Para um bom início, abra um novo terminal e execute:"
Log-Mensagem "INFO" "  harpia --help"
Write-Host ""
Log-Mensagem "SUCESSO" "Finalmente chegamos ao fim"
Log-Mensagem "INFO" "Considere também deixar uma estrelinha no nosso repositório $repo_github"