package main

import (
	"flag"
	"fmt"
	"net/http"
	"sync"
	"time"
)

// Result representa o resultado de uma requisição individual
type Result struct {
	StatusCode int
	Duration   time.Duration
	Error      error
}

// Report contém todas as métricas do teste
type Report struct {
	TotalRequests      int
	SuccessfulRequests int
	FailedRequests     int
	TotalTime          time.Duration
	StatusCodes        map[int]int
	MinDuration        time.Duration
	MaxDuration        time.Duration
	AvgDuration        time.Duration
}

// StressTest representa a configuração do teste de carga
type StressTest struct {
	URL         string
	Requests    int
	Concurrency int
	Client      *http.Client
}

// NewStressTest cria uma nova instância de StressTest
func NewStressTest(url string, requests, concurrency int) *StressTest {
	return &StressTest{
		URL:         url,
		Requests:    requests,
		Concurrency: concurrency,
		Client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// Run executa o teste de carga
func (st *StressTest) Run() *Report {
	results := make(chan Result, st.Requests)
	var wg sync.WaitGroup
	report := &Report{
		StatusCodes: make(map[int]int),
		MinDuration: time.Duration(1<<63 - 1), // Inicializa com o maior valor possível
	}

	// Inicia o timer
	startTime := time.Now()

	// Cria um canal para controlar o número de requests
	requestChan := make(chan struct{}, st.Requests)
	for i := 0; i < st.Requests; i++ {
		requestChan <- struct{}{}
	}
	close(requestChan)

	// Inicia as goroutines de teste
	for i := 0; i < st.Concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for range requestChan {
				start := time.Now()
				resp, err := st.Client.Get(st.URL)
				duration := time.Since(start)

				if err != nil {
					results <- Result{Error: err}
					continue
				}

				resp.Body.Close()
				results <- Result{
					StatusCode: resp.StatusCode,
					Duration:   duration,
				}
			}
		}()
	}

	// Coleta os resultados
	var totalDuration time.Duration
	for i := 0; i < st.Requests; i++ {
		result := <-results
		report.TotalRequests++

		if result.Error == nil {
			report.StatusCodes[result.StatusCode]++
			if result.StatusCode == http.StatusOK {
				report.SuccessfulRequests++
			} else {
				report.FailedRequests++
			}

			// Atualiza métricas de duração
			totalDuration += result.Duration
			if result.Duration < report.MinDuration {
				report.MinDuration = result.Duration
			}
			if result.Duration > report.MaxDuration {
				report.MaxDuration = result.Duration
			}
		} else {
			report.FailedRequests++
		}
	}

	// Calcula o tempo total e a duração média
	report.TotalTime = time.Since(startTime)
	if report.SuccessfulRequests > 0 {
		report.AvgDuration = totalDuration / time.Duration(report.SuccessfulRequests)
	}

	return report
}

func main() {
	// Configuração dos flags
	url := flag.String("url", "", "URL do serviço a ser testado")
	requests := flag.Int("requests", 0, "Número total de requests")
	concurrency := flag.Int("concurrency", 0, "Número de chamadas simultâneas")
	flag.Parse()

	// Validação dos parâmetros
	if *url == "" || *requests <= 0 || *concurrency <= 0 {
		fmt.Println("Erro: Todos os parâmetros são obrigatórios e devem ser válidos")
		fmt.Println("Uso: ./stress-test --url=<URL> --requests=<N> --concurrency=<N>")
		return
	}

	// Cria e executa o teste
	test := NewStressTest(*url, *requests, *concurrency)
	report := test.Run()

	// Imprime o relatório
	printReport(report)
}

func printReport(report *Report) {
	fmt.Println("\n=== Relatório do Teste de Carga ===")
	fmt.Printf("Tempo Total: %v\n", report.TotalTime)
	fmt.Printf("Total de Requests: %d\n", report.TotalRequests)
	fmt.Printf("Requests com Sucesso (200): %d\n", report.SuccessfulRequests)
	fmt.Printf("Requests com Falha: %d\n", report.FailedRequests)

	fmt.Println("\nMétricas de Duração:")
	fmt.Printf("Duração Mínima: %v\n", report.MinDuration)
	fmt.Printf("Duração Máxima: %v\n", report.MaxDuration)
	fmt.Printf("Duração Média: %v\n", report.AvgDuration)

	fmt.Println("\nDistribuição de Status HTTP:")
	for status, count := range report.StatusCodes {
		fmt.Printf("Status %d: %d requests (%.2f%%)\n",
			status,
			count,
			float64(count)/float64(report.TotalRequests)*100)
	}
}
