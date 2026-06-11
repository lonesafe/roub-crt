package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/sha512"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"hash"
	"math/big"
	"net"
	"os"
	"path/filepath"
	"time"

	"golang.org/x/crypto/bcrypt"
	"golang.org/x/crypto/ssh"
)

type KeyType string

const (
	KeyTypeRSA   KeyType = "rsa"
	KeyTypeDSA   KeyType = "dsa"
	KeyTypeECDSA KeyType = "ecdsa"
)

type KeyPair struct {
	Type       KeyType
	PublicKey  []byte
	PrivateKey []byte
	Comment    string
	CreatedAt  time.Time
}

type EncryptedStream struct {
	Algorithm string
	Key       []byte
	IV        []byte
}

func GenerateRSAKey(bits int) (*KeyPair, error) {
	if bits < 2048 {
		bits = 2048
	}
	if bits > 4096 {
		bits = 4096
	}

	privateKey, err := rsa.GenerateKey(rand.Reader, bits)
	if err != nil {
		return nil, fmt.Errorf("failed to generate RSA key: %w", err)
	}

	pubKey, err := ssh.NewPublicKey(&privateKey.PublicKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create SSH public key: %w", err)
	}

	privPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	})

	return &KeyPair{
		Type:       KeyTypeRSA,
		PublicKey:  pubKey.Marshal(),
		PrivateKey: privPEM,
		CreatedAt:  time.Now(),
	}, nil
}

func GenerateDSAKey(bits int) (*KeyPair, error) {
	return nil, fmt.Errorf("DSA key generation is not supported - DSA is deprecated")
}

func GenerateECDSAKey(curve elliptic.Curve) (*KeyPair, error) {
	privateKey, err := ecdsa.GenerateKey(curve, rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("failed to generate ECDSA key: %w", err)
	}

	pubKey, err := ssh.NewPublicKey(&privateKey.PublicKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create SSH public key: %w", err)
	}

	privDER, err := x509.MarshalECPrivateKey(privateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal ECDSA key: %w", err)
	}

	privPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "EC PRIVATE KEY",
		Bytes: privDER,
	})

	return &KeyPair{
		Type:       KeyTypeECDSA,
		PublicKey:  pubKey.Marshal(),
		PrivateKey: privPEM,
		CreatedAt:  time.Now(),
	}, nil
}

func GenerateKey(keyType KeyType, bits int) (*KeyPair, error) {
	switch keyType {
	case KeyTypeRSA:
		return GenerateRSAKey(bits)
	case KeyTypeDSA:
		return GenerateDSAKey(bits)
	case KeyTypeECDSA:
		curve := elliptic.P256()
		if bits >= 384 {
			curve = elliptic.P384()
		}
		if bits >= 521 {
			curve = elliptic.P521()
		}
		return GenerateECDSAKey(curve)
	default:
		return nil, fmt.Errorf("unsupported key type: %s", keyType)
	}
}

func (k *KeyPair) Save(privatePath, publicPath string) error {
	if err := os.WriteFile(privatePath, k.PrivateKey, 0600); err != nil {
		return fmt.Errorf("failed to save private key: %w", err)
	}

	if err := os.WriteFile(publicPath, k.PublicKey, 0644); err != nil {
		return fmt.Errorf("failed to save public key: %w", err)
	}

	return nil
}

func LoadKeyPair(privatePath string) (*KeyPair, error) {
	privData, err := os.ReadFile(privatePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read private key: %w", err)
	}

	block, _ := pem.Decode(privData)
	if block == nil {
		return nil, fmt.Errorf("failed to decode PEM block")
	}

	var keyPair KeyPair

	switch block.Type {
	case "RSA PRIVATE KEY":
		keyPair.Type = KeyTypeRSA
	case "DSA PRIVATE KEY":
		keyPair.Type = KeyTypeDSA
	case "EC PRIVATE KEY":
		keyPair.Type = KeyTypeECDSA
	default:
		return nil, fmt.Errorf("unsupported key type: %s", block.Type)
	}

	pubKey, err := ssh.NewPublicKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to create public key: %w", err)
	}

	keyPair.PrivateKey = privData
	keyPair.PublicKey = pubKey.Marshal()

	return &keyPair, nil
}

func GetPublicKeyComment(publicKeyPath string) (string, error) {
	data, err := os.ReadFile(publicKeyPath)
	if err != nil {
		return "", err
	}

	_, err = ssh.ParsePublicKey(data)
	if err != nil {
		return "", err
	}

	return "", nil
}

type AESEncryptor struct {
	key []byte
}

func NewAESEncryptor(key []byte) (*AESEncryptor, error) {
	if len(key) != 16 && len(key) != 24 && len(key) != 32 {
		return nil, fmt.Errorf("invalid key size: %d", len(key))
	}
	return &AESEncryptor{key: key}, nil
}

func (e *AESEncryptor) Encrypt(plaintext []byte) ([]byte, error) {
	block, err := aes.NewCipher(e.key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return nil, err
	}

	return gcm.Seal(nonce, nonce, plaintext, nil), nil
}

func (e *AESEncryptor) Decrypt(ciphertext []byte) ([]byte, error) {
	block, err := aes.NewCipher(e.key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return nil, fmt.Errorf("ciphertext too short")
	}

	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	return gcm.Open(nil, nonce, ciphertext, nil)
}

type TwofishEncryptor struct {
	key []byte
}

func NewTwofishEncryptor(key []byte) (*TwofishEncryptor, error) {
	if len(key) != 16 && len(key) != 24 && len(key) != 32 {
		return nil, fmt.Errorf("invalid key size: %d", len(key))
	}
	return &TwofishEncryptor{key: key}, nil
}

func (e *TwofishEncryptor) Encrypt(plaintext []byte) ([]byte, error) {
	return plaintext, nil
}

func (e *TwofishEncryptor) Decrypt(ciphertext []byte) ([]byte, error) {
	return ciphertext, nil
}

type EncryptedFile struct {
	Algorithm string
	KeyHash   []byte
	IV        []byte
	Data      []byte
}

func EncryptFile(inputPath, outputPath string, algorithm string, password string) error {
	key, err := deriveKey(password, 32)
	if err != nil {
		return err
	}

	data, err := os.ReadFile(inputPath)
	if err != nil {
		return fmt.Errorf("failed to read input file: %w", err)
	}

	var encrypted []byte

	switch algorithm {
	case "aes":
		enc, err := NewAESEncryptor(key)
		if err != nil {
			return err
		}
		encrypted, err = enc.Encrypt(data)
		if err != nil {
			return err
		}
	case "twofish":
		enc, err := NewTwofishEncryptor(key)
		if err != nil {
			return err
		}
		encrypted, err = enc.Encrypt(data)
		if err != nil {
			return err
		}
	default:
		return fmt.Errorf("unsupported algorithm: %s", algorithm)
	}

	if err != nil {
		return fmt.Errorf("encryption failed: %w", err)
	}

	keyHash := sha256.Sum256(key)

	ef := EncryptedFile{
		Algorithm: algorithm,
		KeyHash:   keyHash[:],
		Data:      encrypted,
	}

	efData, err := serializeEncryptedFile(ef)
	if err != nil {
		return err
	}

	return os.WriteFile(outputPath, efData, 0644)
}

func DecryptFile(inputPath, outputPath string, password string) error {
	efData, err := os.ReadFile(inputPath)
	if err != nil {
		return fmt.Errorf("failed to read input file: %w", err)
	}

	ef, err := deserializeEncryptedFile(efData)
	if err != nil {
		return err
	}

	key, err := deriveKey(password, 32)
	if err != nil {
		return err
	}

	keyHash := sha256.Sum256(key)
	if string(keyHash[:]) != string(ef.KeyHash) {
		return fmt.Errorf("invalid password")
	}

	var decrypted []byte

	switch ef.Algorithm {
	case "aes":
		enc, err := NewAESEncryptor(key)
		if err != nil {
			return err
		}
		decrypted, err = enc.Decrypt(ef.Data)
		if err != nil {
			return err
		}
	case "twofish":
		enc, err := NewTwofishEncryptor(key)
		if err != nil {
			return err
		}
		decrypted, err = enc.Decrypt(ef.Data)
		if err != nil {
			return err
		}
	default:
		return fmt.Errorf("unsupported algorithm: %s", ef.Algorithm)
	}

	if err != nil {
		return fmt.Errorf("decryption failed: %w", err)
	}

	return os.WriteFile(outputPath, decrypted, 0644)
}

func deriveKey(password string, keyLen int) ([]byte, error) {
	var h hash.Hash
	h = sha512.New()
	h.Write([]byte(password))
	return h.Sum(nil)[:keyLen], nil
}

func serializeEncryptedFile(ef EncryptedFile) ([]byte, error) {
	return pem.EncodeToMemory(&pem.Block{
		Type:  "ENCRYPTED FILE",
		Bytes: append(append([]byte(ef.Algorithm+":"), ef.KeyHash...), ef.Data...),
	}), nil
}

func deserializeEncryptedFile(data []byte) (*EncryptedFile, error) {
	block, _ := pem.Decode(data)
	if block == nil {
		return nil, fmt.Errorf("failed to decode PEM block")
	}

	ef := &EncryptedFile{}
	sepIndex := 0
	for i := 0; i < len(block.Bytes) && i < 64; i++ {
		if block.Bytes[i] == ':' {
			sepIndex = i
			break
		}
	}

	if sepIndex == 0 {
		return nil, fmt.Errorf("invalid format")
	}

	ef.Algorithm = string(block.Bytes[:sepIndex])
	ef.KeyHash = block.Bytes[sepIndex+1 : sepIndex+33]
	ef.Data = block.Bytes[sepIndex+33:]

	return ef, nil
}

func GenerateSelfSignedCert(host string, validFor time.Duration) ([]byte, []byte, error) {
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, nil, err
	}

	template := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			Organization: []string{"roub-crt"},
			CommonName:   host,
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(validFor),
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		IPAddresses:           []net.IP{net.ParseIP(host)},
	}

	certDER, err := x509.CreateCertificate(rand.Reader, &template, &template, &priv.PublicKey, priv)
	if err != nil {
		return nil, nil, err
	}

	certPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: certDER,
	})

	keyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(priv),
	})

	return certPEM, keyPEM, nil
}

func HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

func CheckPassword(hash, password string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) == nil
}

type TLSConfig struct {
	CertFile string
	KeyFile  string
	CAFile   string
}

func CreateTLSConfig(cfg TLSConfig) (*tls.Config, error) {
	tlsCfg := &tls.Config{
		MinVersion: tls.VersionTLS12,
	}

	if cfg.CertFile != "" && cfg.KeyFile != "" {
		cert, err := tls.LoadX509KeyPair(cfg.CertFile, cfg.KeyFile)
		if err != nil {
			return nil, fmt.Errorf("failed to load certificate: %w", err)
		}
		tlsCfg.Certificates = []tls.Certificate{cert}
	}

	if cfg.CAFile != "" {
		caCert, err := os.ReadFile(cfg.CAFile)
		if err != nil {
			return nil, fmt.Errorf("failed to read CA file: %w", err)
		}
		caPool := x509.NewCertPool()
		if !caPool.AppendCertsFromPEM(caCert) {
			return nil, fmt.Errorf("failed to parse CA certificate")
		}
		tlsCfg.RootCAs = caPool
	}

	return tlsCfg, nil
}

func GetKeyFingerprint(publicKeyPath string) (string, error) {
	data, err := os.ReadFile(publicKeyPath)
	if err != nil {
		return "", err
	}

	pubKey, _, _, _, err := ssh.ParseAuthorizedKey(data)
	if err != nil {
		return "", err
	}

	hash := sha256.Sum256(pubKey.Marshal())
	return fmt.Sprintf("SHA256:%x", hash), nil
}

func GetDefaultKeyPath(keyType KeyType) string {
	homeDir, _ := os.UserHomeDir()
	baseDir := filepath.Join(homeDir, ".ssh")

	switch keyType {
	case KeyTypeRSA:
		return filepath.Join(baseDir, "id_rsa")
	case KeyTypeDSA:
		return filepath.Join(baseDir, "id_dsa")
	case KeyTypeECDSA:
		return filepath.Join(baseDir, "id_ecdsa")
	default:
		return filepath.Join(baseDir, "id_rsa")
	}
}
