package share

// Comandos permitidos antes do login
const (
	Login  = "login"
	SignUp = "signup"
)

// Comandos permitidos após o login
const (
	Message    = "msg"
	Hidden     = "hidden"
	Users      = "users"
	Logout     = "logout"
	Changenick = "changenick"
)

// Comandos permitidos a qualquer momento
const (
	Help  = "help"
	Exit  = "exit"
	Clear = "clear"
)

// Comandos automáticos usados pelos programas e não pelo usuário
const (
	Fetch    = "fetch"
	Response = "response"
)

const (
	StatusSuccess = 1
	StatusNeutral = 0
	StatusError   = -1
)

// Tamanho de buffers para usuário e servidor
const (
	ClientBuffer = 4096
	ServerBuffer = 512
)

// Mensagens do client
const (
	WelcomeMsg             = " > Olá, bem vindo ao ChatGo, digite algo pra começar." + Reset
	UnexpectMsg            = " > Esse comando não existe, digite novamente. Se tiver dúvidas digite " + Bold + "`help`" + Reset
	AlreadyLoggedInMsg     = " > Você já está logado. Saia primeiro antes de realizar essa ação."
	InvalidLoginCommandMsg = " > Seu comando de login é inválido, revise e tente novamente. Se tiver duvidas, digite " + Bold + "`help`" + Reset
	NotLoggedInMsg         = " > Você não está logado. Realize o login primeiro."
)

// Mensagens do server
const (
	EmptyMessageMsg = " error: mensagem vazia."
)
