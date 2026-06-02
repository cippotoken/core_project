package generator

import (
	"bufio"
	"fmt"
	"os"
	"wallet-generator-scanner/crypto"
)

// GenerateWalletsToFile generates totalCount random keys and writes them directly to filePath using a highly efficient buffered writer.
func GenerateWalletsToFile(filePath string, totalCount int) error {
	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	// Use a large buffer (256KB) to minimize disk IO overhead for 100k writes
	writer := bufio.NewWriterSize(file, 256*1024)
	defer writer.Flush()

	fmt.Printf("[Generator] Generating %d random private keys to %s...\n", totalCount, filePath)

	for i := 1; i <= totalCount; i++ {
		key, err := crypto.GenerateRandomHexKey()
		if err != nil {
			return fmt.Errorf("failed generating random key at index %d: %w", i, err)
		}
		_, err = writer.WriteString("WALLET_PRIVATE_KEY=" + key + "\n")
		if err != nil {
			return err
		}
	}

	fmt.Println("[Generator] Generation completed successfully!")
	return nil
}
