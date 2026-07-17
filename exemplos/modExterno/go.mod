module exemplo.com/externos

go 1.24.2

require github.com/mat-dgruber/Harpia v0.5.0

// Apenas faz com que o Go nao precise baixar as dependecias, ele usa do nosso módulo principal
replace github.com/mat-dgruber/Harpia => ../..
