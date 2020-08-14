package main

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	turkpos "github.com/OzqurYalcin/turkpos/src"
	uuid "github.com/google/uuid"
)

func main() {
	http.HandleFunc("/", view)
	server := http.Server{Addr: ":8080", ReadTimeout: 30 * time.Second, WriteTimeout: 30 * time.Second}
	server.ListenAndServe()
}

func view(w http.ResponseWriter, r *http.Request) {
	api := &turkpos.API{"T"} // "T": test, "P": production
	request := &turkpos.PaymentRequest{}
	request.Body.Payment.G.ClientCode = "10738"    // Müşteri No
	request.Body.Payment.G.ClientUsername = "Test" // Kullanıcı adı
	request.Body.Payment.G.ClientPassword = "Test" // Şifre

	// Ödeme
	transaction := 100.00 // İşlem tutarı
	commission := 0.94    // Komisyon oranı
	installment := 1      // Taksit

	request.Body.Payment.OrderID = uuid.New().String()   // Sipariş numarası
	request.Body.Payment.PosID = 1029                    // 1029: yurtiçi, yurtdışı: 1023
	request.Body.Payment.Security = "3D"                 // "3D": 3dSecure, "NS": NonSecure
	request.Body.Payment.Description = "Açıklama"        // Açıklama
	request.Body.Payment.CardOwner = "Ad soyad"          // Kart sahibi
	request.Body.Payment.CardNumber = "4546711234567894" // Kart numarası
	request.Body.Payment.CardMonth = "12"                // Son kullanma tarihi (Ay)
	request.Body.Payment.CardYear = "2026"               // Son kullanma tarihi (Yıl)
	request.Body.Payment.CardCvc = "000"                 // Kart Cvc Kodu
	request.Body.Payment.GsmNumber = "5554443322"        // Müşteri cep telefonu
	request.Body.Payment.IPAddr = "1.2.3.4"              // Müşteri ip adresi

	request.Body.Payment.GUID = "0c13d406-873b-403b-9c09-a5766840d98c" // Üye işyerine ait anahtar
	request.Body.Payment.CallbackSuccess = "http://localhost:8080/"    // Ödeme başarılı ise dönülecek sayfa
	request.Body.Payment.CallbackError = "http://localhost:8080/"      // Ödeme başarısız ise dönülecek sayfa
	request.Body.Payment.Referer = "http://localhost:8080/"            // Ödeme sayfası

	// Komisyonu müşteri ödeyecek ise :
	request.Body.Payment.Price = strings.ReplaceAll(fmt.Sprintf("%.2f", transaction), ".", ",")
	request.Body.Payment.Amount = strings.ReplaceAll(fmt.Sprintf("%.2f", transaction+(transaction*commission/100)), ".", ",")
	request.Body.Payment.Installment = installment

	// Komisyonu işyeri ödeyecek ise :
	request.Body.Payment.Price = strings.ReplaceAll(fmt.Sprintf("%.2f", transaction-(transaction*commission/100)), ".", ",")
	request.Body.Payment.Amount = strings.ReplaceAll(fmt.Sprintf("%.2f", transaction), ".", ",")
	request.Body.Payment.Installment = installment
	switch r.Method {
	case "GET":
		if r.URL.Path == "/" {
			fmt.Println(request.Body.Payment.OrderID)
			response := api.Payment(request)
			switch response.Body.Payment.Result.Code {
			case 1: // işlem başarılı
				if request.Body.Payment.Security == "3D" {
					fmt.Println(response.Body.Payment.Result.URL)
					http.Redirect(w, r, response.Body.Payment.Result.URL, http.StatusTemporaryRedirect)
				}
				break
			default: // işlem başarısız
				fmt.Println(response.Body.Payment.Result.Code)     // Hata kodu
				fmt.Println(response.Body.Payment.Result.Message)  // Hata mesajı
				fmt.Println(response.Body.Payment.Result.BankCode) // Bankadan dönen kod
				break
			}
		}
		break
	case "POST": // 3D ödeme sonucu
		r.ParseForm()
		dekontID := r.FormValue("TURKPOS_RETVAL_Dekont_ID") // iptal ve iadelerde kullanılan dekont numarası
		fmt.Println(dekontID)
		break
	}
}
