package gomark


func Run(addr string, app App) error {
	return (&app).Run(addr)
}
