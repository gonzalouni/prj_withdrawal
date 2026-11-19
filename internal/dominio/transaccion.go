package dominio

// Operacion identifica el tipo de transacción financiera.
type Operacion string

// Estado de la transacción (modelo OCL {EXITOSA, RECHAZADA} + ERROR).
type Estado string

const (
	OpRetiro Operacion = "RETIRO"
	OpAbono  Operacion = "ABONO"

	EstadoExitosa   Estado = "EXITOSA"
	EstadoRechazada Estado = "RECHAZADA"
	EstadoError     Estado = "ERROR"

	// MensajeSaldoInsuficiente es el texto exigido por US-02.
	MensajeSaldoInsuficiente = "Saldo insuficiente. Intente con un monto menor"
)

// Transaccion es el registro de auditoría de cada intento (log JSON, US-04).
type Transaccion struct {
	ID           string    `json:"id"`
	Fecha        string    `json:"fecha"` // RFC 3339
	Cuenta       string    `json:"cuenta"`
	Tipo         Operacion `json:"tipo"`
	Monto        Dinero    `json:"monto"`
	Estado       Estado    `json:"estado"`
	SaldoAntes   Dinero    `json:"saldo_antes"`
	SaldoDespues Dinero    `json:"saldo_despues"`
	Motivo       string    `json:"motivo"`
}
