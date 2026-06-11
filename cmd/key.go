package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"roub-crt/internal/crypto"
)

var keyType string
var keyBits int
var keyComment string
var keyOutput string

func init() {
	rootCmd.AddCommand(keygenCmd)
	rootCmd.AddCommand(keyListCmd)
	rootCmd.AddCommand(keyFingerprintCmd)

	keygenCmd.Flags().StringVarP(&keyType, "type", "t", "rsa", "Key type (rsa, dsa, ecdsa)")
	keygenCmd.Flags().IntVarP(&keyBits, "bits", "b", 2048, "Key bits (RSA: 2048-4096, DSA: 1024, ECDSA: 256-521)")
	keygenCmd.Flags().StringVarP(&keyComment, "comment", "c", "", "Key comment")
	keygenCmd.Flags().StringVarP(&keyOutput, "output", "o", "", "Output file base name (without extension)")
}

var keygenCmd = &cobra.Command{
	Use:   "keygen",
	Short: "Generate SSH key pairs",
	Long:  `Generate new SSH key pairs for authentication. Supports RSA, DSA, and ECDSA key types.`,
	Run: func(cmd *cobra.Command, args []string) {
		runKeygen()
	},
}

func runKeygen() {
	kt := crypto.KeyType(keyType)

	fmt.Printf("Generating %s key with %d bits...\n", keyType, keyBits)

	keyPair, err := crypto.GenerateKey(kt, keyBits)
	if err != nil {
		fmt.Printf("Error generating key: %v\n", err)
		return
	}

	outputBase := keyOutput
	if outputBase == "" {
		outputBase = crypto.GetDefaultKeyPath(kt)
	}

	privateKeyPath := outputBase
	publicKeyPath := outputBase + ".pub"

	if _, err := os.Stat(filepath.Dir(privateKeyPath)); os.IsNotExist(err) {
		homeDir, _ := os.UserHomeDir()
		sshDir := filepath.Join(homeDir, ".ssh")
		os.MkdirAll(sshDir, 0700)
	}

	if err := keyPair.Save(privateKeyPath, publicKeyPath); err != nil {
		fmt.Printf("Error saving key: %v\n", err)
		return
	}

	fmt.Printf("\nKey pair generated successfully!\n")
	fmt.Printf("Private key: %s\n", privateKeyPath)
	fmt.Printf("Public key:  %s\n", publicKeyPath)

	if keyComment != "" {
		fmt.Printf("Comment:     %s\n", keyComment)
	}

	fingerprint, err := crypto.GetKeyFingerprint(publicKeyPath)
	if err == nil {
		fmt.Printf("Fingerprint: %s\n", fingerprint)
	}

	fmt.Println("\nRemember to add the public key to your ~/.ssh/authorized_keys file on the remote server.")
}

var keyListCmd = &cobra.Command{
	Use:   "list",
	Short: "List SSH keys",
	Run: func(cmd *cobra.Command, args []string) {
		homeDir, _ := os.UserHomeDir()
		sshDir := filepath.Join(homeDir, ".ssh")

		keyTypes := []string{"id_rsa", "id_dsa", "id_ecdsa"}

		fmt.Printf("\n%-30s %-15s %-40s\n", "Key File", "Type", "Fingerprint")
		fmt.Println("------------------------------------------------------------------------")

		for _, keyName := range keyTypes {
			keyPath := filepath.Join(sshDir, keyName)
			pubPath := keyPath + ".pub"

			if _, err := os.Stat(keyPath); err == nil {
				keyPair, err := crypto.LoadKeyPair(keyPath)
				if err != nil {
					continue
				}

				fingerprint, _ := crypto.GetKeyFingerprint(pubPath)
				if fingerprint == "" {
					fingerprint = "N/A"
				}

				fmt.Printf("%-30s %-15s %-40s\n", keyPath, keyPair.Type, fingerprint)
			}
		}
		fmt.Println()
	},
}

var keyFingerprintCmd = &cobra.Command{
	Use:   "fingerprint [key-file]",
	Short: "Get key fingerprint",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		keyPath := args[0]

		fingerprint, err := crypto.GetKeyFingerprint(keyPath)
		if err != nil {
			return fmt.Errorf("failed to get fingerprint: %w", err)
		}

		fmt.Printf("Fingerprint: %s\n", fingerprint)
		return nil
	},
}
