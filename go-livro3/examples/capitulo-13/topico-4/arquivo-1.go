// Programa mínimo: uma pergunta de cada tipo ao servidor System One, só com a
// biblioteca padrão. Aponte DECISOR_URL para o laya-serve (ou para a API do Jev,
// com DECISOR_CHAVE).
package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

func main() {
	base := os.Getenv("DECISOR_URL")
	if base == "" {
		base = "http://127.0.0.1:18690"
	}
	pedido := map[string]any{
		"model": "multilingual",
		"state": "Fui cobrado duas vezes em março. Estorne a cobrança duplicada hoje ou cancelo o plano.",
		"questions": map[string]any{
			"area": map[string]any{
				"type":         "choice",
				"instructions": "Qual área deve atender este chamado?",
				"criteria":     map[string]any{"cobranca": "faturas e estornos", "tecnico": "erros e quedas", "outro": nil},
			},
			"urgencia": map[string]any{
				"type":         "score",
				"instructions": "Qual a urgência?",
				"criteria":     []string{"baixa", "média", "alta"},
			},
			"churn": map[string]any{"type": "noul", "instructions": "O usuário ameaça cancelar o plano?"},
		},
	}
	corpo, _ := json.Marshal(pedido)
	req, _ := http.NewRequest(http.MethodPost, base+"/v1/systemone", bytes.NewReader(corpo))
	req.Header.Set("Content-Type", "application/json")
	if k := os.Getenv("DECISOR_CHAVE"); k != "" {
		req.Header.Set("Authorization", "Bearer "+k)
	}
	t0 := time.Now()
	resp, err := (&http.Client{Timeout: 30 * time.Second}).Do(req)
	if err != nil {
		fmt.Println("erro:", err)
		os.Exit(1)
	}
	defer resp.Body.Close()
	dados, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		fmt.Printf("HTTP %d: %s\n", resp.StatusCode, dados)
		os.Exit(1)
	}
	var saida struct {
		Answers map[string]struct {
			Type          string             `json:"type"`
			Choice        string             `json:"choice"`
			Score         float64            `json:"score"`
			Noul          float64            `json:"noul"`
			Probabilities map[string]float64 `json:"probabilities"`
			Confidence    float64            `json:"confidence"`
		} `json:"answers"`
	}
	if err := json.Unmarshal(dados, &saida); err != nil {
		fmt.Println("resposta fora do contrato:", err)
		os.Exit(1)
	}
	a := saida.Answers
	fmt.Printf("área: %s (confiança %.3f)\n", a["area"].Choice, a["area"].Confidence)
	fmt.Printf("urgência: %.2f na escala 0 a 2\n", a["urgencia"].Score)
	fmt.Printf("risco de cancelamento: P=%.3f\n", a["churn"].Noul)
	fmt.Printf("uma chamada, três perguntas, %v\n", time.Since(t0).Round(time.Millisecond))
}
