package libs

import (
    "encoding/json"
    "time"

    "github.com/katena-chain/sdk-go-client/api"
    "github.com/katena-chain/sdk-go-client/client"
    "github.com/katena-chain/sdk-go-client/crypto/X25519"
    "github.com/katena-chain/sdk-go-client/entity"
    entityApi "github.com/katena-chain/sdk-go-client/entity/api"
    "github.com/katena-chain/sdk-go-client/entity/certify"
    "github.com/katena-chain/sdk-go-client/utils"
)

type Config struct {
    PrivKey        string
    ChainID        string
    CompanyChainID string
    ApiUrl         string
}

type CertificateHandler struct {
    Config
    UuidText      string
    SignatureText string
    SignerText    string
}

func ConvertKeys(recipientPubKey string, senderPubKey string, senderPrivKey string) (*X25519.PublicKey, *X25519.PublicKey, *X25519.PrivateKey, error) {
    // Converts & returns base64 -> X25519 objects keys

    recipientPublicKey, err := utils.CreatePublicKeyX25519FromBase64(recipientPubKey)
    if err != nil {
        return nil, nil, nil, err
    }
    senderPublicKey, err := utils.CreatePublicKeyX25519FromBase64(senderPubKey)
    if err != nil {
        return nil, nil, nil, err
    }
    senderPrivateKey, err := utils.CreatePrivateKeyX25519FromBase64(senderPrivKey)
    if err != nil {
        return nil, nil, nil, err
    }

    return recipientPublicKey, senderPublicKey, senderPrivateKey, nil
}

func (certHandler *CertificateHandler) GetCertificatePreview() (string, error) {
    // Calls the API to retrieve the certificate corresponding to the structs data and returns the indented JSON result

    privateKeyForTransactor, err := utils.CreatePrivateKeyED25519FromBase64(certHandler.Config.PrivKey)
    if err != nil {
        return "Error loading", err
    }

    // Build the certificate
    certificate := certify.NewCertificateV1(certHandler.UuidText, certHandler.Config.CompanyChainID, []byte(certHandler.SignatureText), []byte(certHandler.SignerText))
    message := &certify.MsgCreateCertificate{
        Certificate: certificate,
    }

    // getTransaction() reproduction here
    nonceTime := entity.Time{
        Time: time.Now(),
    }
    sealState := &entity.SealState{
        Message:   message,
        ChainID:   certHandler.Config.ChainID,
        NonceTime: &nonceTime,
    }
    sealStateBytes, err := sealState.GetSignBytes()
    if err != nil {
        return "Error loading", err
    }

    msgSignature := privateKeyForTransactor.Sign(sealStateBytes)
    transaction := entityApi.NewTransaction(message, msgSignature, privateKeyForTransactor.GetPublicKey(), &nonceTime)

    // Indent the json corresponding to the transaction
    data, err := json.MarshalIndent(transaction, " ", "    ")
    if err != nil {
        return "Error loading", err
    }

    return string(data), nil
}

func (certHandler *CertificateHandler) SendCertificate() (*entityApi.TransactionStatus, error) {
    // Sends a certificate to the API using the data in the built struct

    privateKeyForTransactor, err := utils.CreatePrivateKeyED25519FromBase64(certHandler.Config.PrivKey)
    if err != nil {
        return nil, err
    }

    // Build the transactor
    transactor := client.NewTransactor(certHandler.Config.ApiUrl, certHandler.Config.ChainID, privateKeyForTransactor, certHandler.Config.CompanyChainID)

    // Send the transaction and display the resulting code and message in the last tab
    transactionStatus, err := transactor.SendCertificateV1(certHandler.UuidText, []byte(certHandler.SignatureText), []byte(certHandler.SignerText))
    if err != nil {
        return nil, err
    }

    return transactionStatus, nil
}

func (certHandler *CertificateHandler) RetrieveCertificate() (string, error) {
    // Create transactor & retrieve certificate with the data in the struct

    apiHandler := api.NewHandler(certHandler.Config.ApiUrl)
    transactionWrapper, err := apiHandler.RetrieveCertificate(certHandler.Config.CompanyChainID, certHandler.UuidText)
    if err != nil {
        return "", err
    }

    // Indent and display the json corresponding to the transaction
    data, err := json.MarshalIndent(transactionWrapper, " ", "    ")
    if err != nil {
        return "", err
    }

    return string(data), nil
}

type SecretHandler struct {
    Config
    UuidText        string
    Content         []byte
    RecipientPubKey *X25519.PublicKey
    SenderPubKey    *X25519.PublicKey
    SenderPrivKey   *X25519.PrivateKey
}

func (secHandler *SecretHandler) GetSecretPreview() (string, error) {
    // Calls the API to retrieve the secret using the struct's data and returns the indented JSON result

    privateKeyForTransactor, err := utils.CreatePrivateKeyED25519FromBase64(secHandler.Config.PrivKey)
    if err != nil {
        return "Error loading", err
    }

    // Encrypt the secret
    nonce, encryptedContent, err := secHandler.SenderPrivKey.Seal(secHandler.Content, secHandler.RecipientPubKey)
    if err != nil {
        return "Error loading", err
    }

    // Getting the transaction
    secret := certify.NewSecretV1(
        secHandler.UuidText,
        secHandler.Config.CompanyChainID,
        secHandler.SenderPubKey,
        nonce,
        encryptedContent,
    )
    messageSecret := &certify.MsgCreateSecret{
        Secret: secret,
    }
    nonceTime := entity.Time{
        Time: time.Now(),
    }
    sealState := &entity.SealState{
        Message:   messageSecret,
        ChainID:   secHandler.Config.ChainID,
        NonceTime: &nonceTime,
    }
    sealStateBytes, err := sealState.GetSignBytes()
    if err != nil {
        return "Error loading", err
    }

    msgSignature := privateKeyForTransactor.Sign(sealStateBytes)
    transaction := entityApi.NewTransaction(messageSecret, msgSignature, privateKeyForTransactor.GetPublicKey(), &nonceTime)

    // Indent and display the json corresponding to the transaction
    data, err := json.MarshalIndent(transaction, " ", "    ")
    if err != nil {
        return "Error loading", err
    }

    return string(data), nil
}

func (secHandler *SecretHandler) SendSecret() (*entityApi.TransactionStatus, error) {
    // Sends a secret to the API using the data in the struct

    privateKeyForTransactor, err := utils.CreatePrivateKeyED25519FromBase64(secHandler.Config.PrivKey)
    if err != nil {
        return nil, err
    }

    transactorSecret := client.NewTransactor(secHandler.Config.ApiUrl, secHandler.Config.ChainID, privateKeyForTransactor, secHandler.Config.CompanyChainID)
    // Encrypt the secret (again to refresh nonce)
    nonce, encryptedContent, err := secHandler.SenderPrivKey.Seal(secHandler.Content, secHandler.RecipientPubKey)
    if err != nil {
        return nil, err
    }
    transactionStatusSecret, err := transactorSecret.SendSecretV1(secHandler.UuidText, secHandler.SenderPubKey, nonce, encryptedContent)
    if err != nil {
        return nil, err
    }

    return transactionStatusSecret, nil
}
