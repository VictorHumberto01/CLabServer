package banner

import "fmt"

// PrintBanner displays the cLab ASCII art banner
func PrintBanner() {
	banner := `
    ____          _       ____                           
 / ___| |    __ _| |__   / ___|  ___ _ ____   _____ _ __ 
| |   | |   / _' | '_ \  \___ \ / _ \ '__\ \ / / _ \ '__|
| |___| |__| (_| | |_) |  ___) |  __/ |   \ V /  __/ |   
 \____|_____\__,_|_.__/  |____/ \___|_|    \_/ \___|_|    


  
                                
  C Language Learning Assistant
  ============================
`
	fmt.Println(banner)
}
