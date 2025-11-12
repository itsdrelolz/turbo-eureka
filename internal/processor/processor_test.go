package processor

import (
	"fmt"
	"os"
	"testing"
)

func TestExtractText(t *testing.T) {

	data, err := os.ReadFile("Andre_Harper_Resume.pdf")

	if err != nil {
		t.Error("Failed to open file")
	}

	result, err := ExtractTextFromPDF(data)

	if err != nil {
		t.Error("failed to read file")
	}

	fmt.Println(result)
}
