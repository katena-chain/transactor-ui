//go:generate go-bindata -pkg assets -o ../assets/assets.go ../assets/
package main

import (
    "encoding/json"
    "strconv"
    "time"

    "fyne.io/fyne"
    "fyne.io/fyne/app"
    "fyne.io/fyne/dialog"
    "fyne.io/fyne/layout"
    "fyne.io/fyne/theme"
    "fyne.io/fyne/widget"
    "github.com/google/uuid"
    "github.com/katena-chain/sdk-go-client/api"
    _ "github.com/mattn/go-sqlite3"

    "github.com/katena-chain/transactor-ui/libs"
)

func main() {
    // Get a data access object to manipulate the database
    databaseDAO := libs.InitDb()

    // Variable declarations
    var tabCont *widget.TabContainer
    var selectWidgetCertificates *widget.Select
    var selectWidgetSecrets *widget.Select
    var entryDisplaySecrets *widget.Entry

    var config libs.Config
    var certificateData libs.CertificateHandler
    var secretData libs.SecretHandler

    // Generate window
    appl := app.New()
    window := appl.NewWindow("Katena Transactor UI")
    appl.Settings().SetTheme(theme.LightTheme())

    // Preparing icons
    configIcon, err := libs.MakeImageResource("Config icon", "../assets/config.png")
    libs.CheckIcon(err)
    transactionIcon, err := libs.MakeImageResource("Transaction icon", "../assets/transaction.png")
    libs.CheckIcon(err)
    resultIcon, err := libs.MakeImageResource("Result icon", "../assets/result.png")
    libs.CheckIcon(err)
    appIcon, err := libs.MakeImageResource("App icon", "../assets/katena-icon.png")
    libs.CheckIcon(err)

    // Build tabs for tabContainer
    privKeyEntry := widget.NewEntry()
    companyChainIDEntry := widget.NewEntry()
    chainIDEntry := widget.NewEntry()
    chainIDEntry.SetText("katena-chain-test")
    apiURLEntry := widget.NewEntry()
    apiURLEntry.SetText("https://api.test.katena.transchain.io/api/v1")
    tabConfig := widget.NewVBox(
        widget.NewLabel("Company chain id :"),
        companyChainIDEntry,
        widget.NewLabel("Private key :"),
        privKeyEntry,
        widget.NewLabel("Chain id :"),
        chainIDEntry,
        widget.NewLabel("API URL :"),
        apiURLEntry,
        widget.NewButton("Confirm", func() {
            // Opens transactions tab & gets entered the values
            // TODO - verify valid input ?

            config = libs.Config{
                PrivKey:        privKeyEntry.Text,
                CompanyChainID: companyChainIDEntry.Text,
                ChainID:        chainIDEntry.Text,
                ApiUrl:         apiURLEntry.Text,
            }

            tabCont.SelectTabIndex(1)
        }),
    )

    entryDisplayCertificates := widget.NewMultiLineEntry()
    selectWidgetCertificates = widget.NewSelect(
        databaseDAO.UpdateCertificateOptions(), func(optionSelected string) {
            if optionSelected == "Add certificate..." {
                // Adding new uuid

                // Build dialog canvas
                uuidEntry := widget.NewEntry()
                signatureEntry := widget.NewEntry()
                signerEntry := widget.NewEntry()
                dialogContent := widget.NewVBox(
                    widget.NewLabel("UUID :"),
                    uuidEntry,
                    widget.NewButton("Generate UUID", func() {
                        genUUID, err := uuid.NewRandom()
                        if err != nil {
                            return
                        }
                        uuidEntry.SetText(genUUID.String())
                    }),
                    widget.NewLabel("Signature :"),
                    signatureEntry,
                    widget.NewLabel("Signer :"),
                    signerEntry,
                )

                // Build child dialog canvas
                jsonZone := widget.NewMultiLineEntry()
                childDialogContent := widget.NewVBox(
                    widget.NewLabel("Expected transaction :"),
                    jsonZone,
                )

                dialog.ShowCustomConfirm("Add certificate...", "Confirm", "Cancel", dialogContent,
                    func(confirm bool) {
                        // If confirm, we prepare the certificate and ask for confirmation
                        if confirm && uuidEntry.Text != "" && signatureEntry.Text != "" && signerEntry.Text != "" {

                            // Prepare data for the certificate
                            certificateData = libs.CertificateHandler{
                                Config:        config,
                                UuidText:      uuidEntry.Text,
                                SignatureText: signatureEntry.Text,
                                SignerText:    signerEntry.Text,
                            }

                            // Get a JSON preview of the certificate
                            previewData, err := certificateData.GetCertificatePreview()
                            if err != nil {
                                dialog.ShowError(err, window)
                                return
                            }
                            // Display it in the dedicated zone
                            jsonZone.SetText(previewData)

                            dialog.ShowCustomConfirm("Confirm certificate...", "Send certificate", "Cancel", childDialogContent,
                                func(confirm bool) {
                                    // If confirms, save the certificate and send the transaction
                                    if confirm {

                                        // Send the transaction
                                        transactionStatus, err := certificateData.SendCertificate()
                                        if err != nil {
                                            dialog.ShowError(err, window)
                                            return
                                        }

                                        // Add the certificate to the DB
                                        databaseDAO.AddCertificateEntry(certificateData.UuidText, certificateData.SignatureText, certificateData.SignerText)
                                        // Updates the options list & select the created uuid
                                        selectWidgetCertificates.Options = databaseDAO.UpdateCertificateOptions()
                                        window.Canvas().Refresh(selectWidgetCertificates)
                                        selectWidgetCertificates.SetSelected(uuidEntry.Text)

                                        // Show results dialog
                                        // => Convert uint32 statuscode to string
                                        dialog.ShowInformation("Transaction status :", "Transaction code : "+strconv.FormatUint(uint64(transactionStatus.Code), 10)+
                                            "\nTransaction message : "+transactionStatus.Message, window)

                                    }
                                }, window)
                        }
                    }, window)
                return
            }

            // From here - if an already existing UUID is selected
            // Get information about the certificate from the API and display it in the textEntry

            certificateData = libs.CertificateHandler{
                Config:   config,
                UuidText: selectWidgetCertificates.Selected,
            }

            resultData, err := certificateData.RetrieveCertificate()
            if err != nil {
                dialog.ShowError(err, window)
                return
            }

            entryDisplayCertificates.SetText(resultData)
        },
    )

    // The scrollcontainer has to be wrapped in a fixed grid layout in order to be displayed in the proper size
    entryDisplayCertificatesWrapper := widget.NewScrollContainer(entryDisplayCertificates)
    size := fyne.Size{Width: 1000, Height: 700}
    entryCertificatesWrap := fyne.NewContainerWithLayout(layout.NewFixedGridLayout(size), entryDisplayCertificatesWrapper)

    tabCertificates := widget.NewVBox(
        selectWidgetCertificates,
        entryCertificatesWrap,
        widget.NewButton("Remove this certificate", func() {
            // Removes the selected certificate

            databaseDAO.RemoveCertificate(selectWidgetCertificates.Selected)
            // Updates the options list & selected the created certificate
            selectWidgetCertificates.Options = databaseDAO.UpdateCertificateOptions()
            window.Canvas().Refresh(selectWidgetCertificates)
            // Selects first option by default
            selectWidgetCertificates.SetSelected(databaseDAO.UpdateCertificateOptions()[0])
        }),
    )

    selectWidgetSecrets = widget.NewSelect(
        databaseDAO.UpdateSecretOptions(),
        func(optionSelected string) {
            if optionSelected == "Add secret..." {
                // Adding a secret

                // Preparing dialog canvas
                uuidEntrySecrets := widget.NewEntry()
                contentEntry := widget.NewEntry()
                recipientPublicEntry := widget.NewEntry()
                recipientPrivateEntry := widget.NewEntry()
                senderPublicEntry := widget.NewEntry()
                senderPrivateEntry := widget.NewEntry()
                dialogContentSecrets := widget.NewVBox(
                    widget.NewLabel("UUID :"),
                    uuidEntrySecrets,
                    widget.NewButton("Generate UUID", func() {
                        genUUID, err := uuid.NewRandom()
                        if err != nil {
                            return
                        }
                        uuidEntrySecrets.SetText(genUUID.String())
                    }),
                    widget.NewLabel("Content :"),
                    contentEntry,
                    widget.NewLabel("Recipient public key :"),
                    recipientPublicEntry,
                    widget.NewLabel("Recipient private key :"),
                    recipientPrivateEntry,
                    widget.NewLabel("Sender public key :"),
                    senderPublicEntry,
                    widget.NewLabel("Sender private key :"),
                    senderPrivateEntry,
                )

                // Build child dialog canvas
                jsonZoneSecrets := widget.NewMultiLineEntry()
                secretsChildDialogContent := widget.NewVBox(
                    widget.NewLabel("Expected transaction :"),
                    jsonZoneSecrets,
                )

                dialog.ShowCustomConfirm("Add a secret...", "Confirm", "Cancel", dialogContentSecrets, func(confirm bool) {
                    if confirm && uuidEntrySecrets.Text != "" && contentEntry.Text != "" && recipientPublicEntry.Text != "" && recipientPrivateEntry.Text != "" && senderPrivateEntry.Text != "" && senderPublicEntry.Text != "" {
                        // If confirmed, prepare the secret and ask for confirmation

                        // For the db
                        recipientPrivateKeyX25519Base64 := recipientPrivateEntry.Text

                        // Prepare the keys in the right format
                        recipientPublicKey, senderPublicKey, senderPrivateKey, err := libs.ConvertKeys(recipientPublicEntry.Text, senderPublicEntry.Text, senderPrivateEntry.Text)
                        if err != nil {
                            dialog.ShowError(err, window)
                            return
                        }

                        // Get needed values
                        secretData = libs.SecretHandler{
                            Config:          config,
                            UuidText:        uuidEntrySecrets.Text,
                            Content:         []byte(contentEntry.Text),
                            RecipientPubKey: recipientPublicKey,
                            SenderPubKey:    senderPublicKey,
                            SenderPrivKey:   senderPrivateKey,
                        }

                        // Convert keys and get the preview json
                        previewData, err := secretData.GetSecretPreview()
                        if err != nil {
                            dialog.ShowError(err, window)
                            return
                        }

                        // Display the preview
                        jsonZoneSecrets.SetText(previewData)

                        dialog.ShowCustomConfirm("Confirm secret", "Send secret", "Cancel", secretsChildDialogContent, func(confirm bool) {
                            // If confirmed, save the secret to DB and send it to the API
                            if confirm {
                                // API send
                                transactionStatusSecret, err := secretData.SendSecret()
                                if err != nil {
                                    dialog.ShowError(err, window)
                                    return
                                }

                                // DB save
                                databaseDAO.AddSecretEntry(secretData.UuidText, recipientPrivateKeyX25519Base64)
                                // Updates the options list & select the created uuid
                                selectWidgetSecrets.Options = databaseDAO.UpdateSecretOptions()
                                window.Canvas().Refresh(selectWidgetSecrets)

                                // If we don't wait, the selection prompts an error because the secret can't be retrieved from the API yet
                                // Done in a goroutine in order to not freeze the UI for a few seconds
                                go func() {
                                    // TODO - Remove and not select
                                    time.Sleep(2 * time.Second)
                                    selectWidgetSecrets.SetSelected(secretData.UuidText)
                                }()

                                // Show results dialog
                                // => Convert uint32 statuscode to string
                                dialog.ShowInformation("Transaction status :", "Transaction code : "+strconv.FormatUint(uint64(transactionStatusSecret.Code), 10)+
                                    "\nTransaction message : "+transactionStatusSecret.Message, window)

                            }

                        }, window)
                    }
                }, window)
            } else {
                // An already existing secret has been selected
                secretUUID := selectWidgetSecrets.Selected
                // Done in a goroutine because this can return data of an important size which can freeze the UI for a little while if done
                // on the main routine
                go func() {
                    // Retrieve corresponding secret
                    apiHandler := api.NewHandler(config.ApiUrl)
                    transactionWrappers, err := apiHandler.RetrieveSecrets(config.CompanyChainID, secretUUID)
                    if err != nil {
                        dialog.ShowError(err, window)
                        return
                    }

                    // Indent and display the json corresponding to the transaction
                    data, err := json.MarshalIndent(transactionWrappers, " ", "    ")
                    if err != nil {
                        dialog.ShowError(err, window)
                        return
                    }

                    entryDisplaySecrets.SetText(string(data))
                }()
            }
        },
    )

    // The scrollcontainer has to be wrapped in a fixed grid layout in order to be displayed in the proper size
    entryDisplaySecrets = widget.NewMultiLineEntry()
    entryDisplaySecretsWrapper := widget.NewScrollContainer(entryDisplaySecrets)
    size = fyne.Size{Height: 1000, Width: 700}
    entrySecretsWrap := fyne.NewContainerWithLayout(layout.NewFixedGridLayout(size), entryDisplaySecretsWrapper)

    tabSecrets := widget.NewVBox(
        selectWidgetSecrets,
        entrySecretsWrap,
        widget.NewButton("Remove this secret", func() {
            // Removes the selected secret
            databaseDAO.RemoveSecret(selectWidgetSecrets.Selected)
            // Updates the options list & selected the created certificate
            selectWidgetSecrets.Options = databaseDAO.UpdateSecretOptions()
            window.Canvas().Refresh(selectWidgetSecrets)
            // Selects first option by default
            selectWidgetSecrets.SetSelected(databaseDAO.UpdateSecretOptions()[0])
        }),
    )

    entryDisplaySecretsWrapper.Resize(fyne.NewSize(500, 2000))

    // Build tabContainer
    tabCont = widget.NewTabContainer(
        widget.NewTabItemWithIcon("Configuration", configIcon, tabConfig),
        widget.NewTabItemWithIcon("Certificates", transactionIcon, tabCertificates),
        widget.NewTabItemWithIcon("Secrets", resultIcon, tabSecrets),
    )
    tabCont.SetTabLocation(widget.TabLocationLeading)

    window.Resize(fyne.NewSize(1100, 700))
    window.SetIcon(appIcon)
    window.SetContent(widget.NewVBox(
        tabCont,
    ))
    window.CenterOnScreen()
    window.ShowAndRun()
}
