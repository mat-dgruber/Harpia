package ptst

// Objeto é a interface primordial e base de toda a tipagem dinâmica do Portuscript.
//
// Qualquer estrutura em Go que deseje ser interpretada, manipulada ou armazenada
// como uma variável no runtime da máquina virtual do Portuscript deve implementar esta interface.
//
// Ela exige um único método Tipo(), que retorna o ponteiro para a classe estrutural (Tipo) do objeto,
// permitindo a reflexão e resolução dinâmica de métodos e atributos em tempo de execução.
type Objeto interface {
	// Tipo devolve a representação da classe (metadados de Tipo) correspondente à instância.
	Tipo() *Tipo
}
