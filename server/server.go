package main

import (
	"context"
	"fmt"
	"math/rand"
	"net/http"
	"time"
)

// Abrindo a possibilidade para o cancelamento
func VerySlowFunction(ctx context.Context, chSuccess chan bool) {
	seed1 := rand.NewSource(time.Now().UnixNano())
	r := rand.New(seed1)
	duration := r.Intn(10) + 1

	for i := 1; i <= duration; i++ {
		time.Sleep(1 * time.Second)

		select {
		case <-ctx.Done(): // Sinal de cancelamento
			fmt.Printf("Cancelei após %d segundos\n", i)
			return
		default: // Impede de bloquear eternamente enquanto espera um sinal
		}
	}

	fmt.Printf("%d segundos se passaram-se, sessereressessê\n", duration)
	chSuccess <- true
}

func Process(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		// Abrimos mão do Gorilla/Mux: simplificamos lá, pagamos aqui.
		// O browser envia um OPTIONS, então vamos ignorá-lo e ficar com o GET.
		return
	}

	/*
		TESTE:

		Mudando o tempo limite para *menos* do que a função lenta dura. Nesse caso,
		a goroutine continua rodando, o que pode ser o desejado porém, como o Akita
		disse em um vídeo, fazer assíncrono *dentro* do processo não garante a
		retentativa caso o processo caia.

		Vamos supor outra situação: o tempo limite é o limite e, se a função lenta
		não conseguir finalizar nesse tempo, temos de devolver um erro.
	*/

	// Fazendo como no outro exemplo dado na aula: contexto com cancelamento
	reqCtx := r.Context()
	cancelCtx, cancelFn := context.WithCancel(reqCtx)
	successCh := make(chan bool)
	go VerySlowFunction(cancelCtx, successCh)

	select {
	// Novo tempo limite de 5 segundos
	case <-time.After(5 * time.Second):
		cancelFn()
		w.Write([]byte("Tempo limite esgotado"))
	case <-reqCtx.Done():
		cancelFn()
		fmt.Println("Cancelado pelo usuário")
	case <-successCh:
		cancelFn() // Só para agradar compilador
		w.Write([]byte("OK"))
	}

}

func main() {
	http.ListenAndServe(":8080", http.HandlerFunc(Process))
}
