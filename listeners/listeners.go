package listeners

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"sync"
	"time"

	"listener-toolkit/database"

	"github.com/gorilla/mux"
	"github.com/tarm/serial"
)

var mutex = &sync.Mutex{}

func uploader(body io.Reader) {

	mutex.Lock()

	client := &http.Client{}
	req, err := http.NewRequest("POST", "https://dicom.sean.ph", body)
	if err != nil {
		log.Fatal(err.Error())
	}
	req.Header.Set("Content-Type", "application/json")

	res, err := client.Do(req)

	if err != nil {
		log.Fatal(err.Error())
	}

	defer res.Body.Close()

	database.InsertData("")

	mutex.Unlock()

}

type ResponseMessage struct {
	Success bool   `json:"success"`
	Code    int    `json:"code,omitempty"`
	Message string `json:"message"`
}

func (rm *ResponseMessage) SuccessResponse(message string, code int) []byte {
	rm.Success = true
	rm.Code = code
	rm.Message = message

	jsonString, _ := json.Marshal(rm)
	return jsonString
}

func (rm *ResponseMessage) ErrorResponse(message string, code int) []byte {
	rm.Success = false
	rm.Code = code
	rm.Message = message

	jsonString, _ := json.Marshal(rm)
	return jsonString
}

func ecmHandler(w http.ResponseWriter, r *http.Request) {
	mutex.Lock()
	w.Header().Set("Content-Type", "text/xml")
	w.WriteHeader(202)
	w.Write([]byte(`
		<?xml version="1.0" encoding="UTF-8"?>
		<!DOCTYPE ecg-response SYSTEM "ecg-response.dtd">
		<ecg-response>
			<response-code>0</response-code>
			<response-message>Success</response-message>
		</ecg-response>
	`))
	mutex.Unlock()
}

func dicomHandler(w http.ResponseWriter, r *http.Request) {
	mutex.Lock()
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)

	response := ResponseMessage{}
	jsonResponse := response.SuccessResponse("Success", 0)
	w.Write(jsonResponse)
	mutex.Unlock()
}

func handleTCPConnection(conn net.Conn) {
	defer conn.Close()

	addr := conn.RemoteAddr().String()

	buffer := make([]byte, 1024)

	n, err := conn.Read(buffer)
	if err != nil {
		log.Println(err.Error())
	}
	log.Println(addr+":", string(buffer[:n]))

	reply := buffer[:n]
	conn.Write([]byte(reply))
	conn.Close()
}

func InitializeTCP(addr string, port string) {

	log.Println("Initializing TCP server on port " + port)
	server, err := net.Listen("tcp", addr+":"+port)
	if err != nil {
		log.Fatal(err.Error())
	}
	defer server.Close()

	log.Println("TCP server listening on port " + port)

	for {
		conn, err := server.Accept()
		if err != nil {
			log.Fatal(err.Error())
		}
		go handleTCPConnection(conn)
	}
}

func InitializeUDP(addr string, port string) {

	log.Println("Initializing UDP server on port " + port)

	server, err := net.ListenPacket("udp", addr+":"+port)

	if err != nil {
		log.Fatal(err.Error())
	}
	defer server.Close()

	log.Println("UDP server listening on port " + port)

	for {
		buffer := make([]byte, 1024)
		n, addr, err := server.ReadFrom(buffer)

		if err != nil {
			log.Fatal(err.Error())
		}
		log.Println(addr.String()+":", string(buffer[:n]))

		reply := buffer[:n]
		server.WriteTo(reply, addr)
	}
}

func InitializeHTTP(port string) {

	r := mux.NewRouter()
	r.HandleFunc("/ems", ecmHandler) // .Methods("GET", "POST", "OPTIONS")
	r.HandleFunc("/dicom", dicomHandler)
	r.Use(mux.CORSMethodMiddleware(r))

	serve := &http.Server{
		Handler:      r,
		Addr:         ":" + port,
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}
	go serve.ListenAndServe()
	log.Println("HTTP server listening on port " + port)
	log.Println("DICOM handler available at /dicom")
	log.Println("EMS available at /ems")
}

func InitializeSerial(path string, baud int) {

	log.Println("Initializing serial connection to " + path + " at " + fmt.Sprint(baud) + " baud")
	config := &serial.Config{
		Name: path,
		Baud: baud,
	}

	serial, err := serial.OpenPort(config)
	if err != nil {
		log.Println("Unable to locate " + path)
		return
	}

	defer serial.Close()

	buf := make([]byte, 1024)

	for {
		_, err := serial.Write([]byte("test"))

		if err != nil {
			log.Println("Unable to write to serial port")
			return
		}

		// log.Println("Wrote " + fmt.Sprint(n) + " bytes")

		n, err := serial.Read(buf)
		if err != nil {
			log.Println(err)
		}

		fmt.Println(string(buf[:n]))
	}

}
