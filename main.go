package main

import (
    "log"
	"fmt"
	"net/http"
    "io/ioutil"

    "github.com/stellar/go/keypair"
    "github.com/stellar/go/clients/horizon"	
	"github.com/stellar/go/hash"
	b "github.com/stellar/go/build"
)

type NamedAccount struct {
	Address string
	Account horizon.Account
}
type SigningEntry struct {
	Key string
	weight uint32
}

func ClientEnv() *horizon.Client { return  horizon.DefaultTestNetClient }

func main() {
	
	var asset_name string = "MyGroupedSec" // Change this to the Name of the Asset to securitize
	
	var issuer_passphrase = "Issuer:I am the Asset Issuer!" // Change this to the issuer account!
	// SDYPFDJH5AD7UFLLU3OX334YWNCDPOXFZIFAE4DIE72F3WSE76R2BYJQ
	// GCBVL3SQFEVKRLP6AJ47UKKWYEBY45WHAJHCEZKVWMTGMCT4H4NKQYLH
	var issuer_pair = MakePair(issuer_passphrase)
		
	var recipient_passphrase = "Recipient:I am the Asset Holder!" // Change this to an account that "holds" the asset!
	// SAFP43MHLANLI3QWF6F65BB4FTKACFSTOM7W7H7R6GAXRWQYYRAK5Q52
	// GBB3EEK3KKPA7Z4NDW4ZZ5YKCRT5S7RTFZ5XIXHHNWNRLR2K77H6AFR3
	var recipient_pair = MakePair(recipient_passphrase)
	
	var issue_passphrase = "AssetRep:" + issuer_pair.Address() + "/" + asset_name
	// SAGK3AX7FGPIJ7N5SPLVAQA7SIPTPQ7LQIIVFVQDAWKF5A4E6XCHIXBY
	// GBFT3FUYMS5Z5HUKPFENTHFJXUAAKGELGHJBLKBGX3X45OSDXVGZFMZB
	var issue_pair = MakePair(issue_passphrase)

	FundAccount(issuer_pair)

	
	FundAccount(recipient_pair)
	set_trust(true, recipient_pair.Seed(), recipient_pair.Seed(), asset_name, issuer_pair.Address())
	//set_trust(false, recipient_pair.Seed(), recipient_pair.Seed(), asset_name, issuer_pair.Address())
	add_signer(issuer_pair.Seed(), issuer_pair.Seed(), recipient_pair.Address(), uint32(1))
	
	
	issuer_acct := LoadAccount(issuer_pair.Address())
	FundAccount(issue_pair)
	for _, signer := range issuer_acct.Account.Signers {
		add_signer(issue_pair.Seed(), issue_pair.Seed(), signer.PublicKey, uint32(signer.Weight))
	}
	master_weight(issue_pair.Seed(), issue_pair.Seed(), 0)

	test_tx(issuer_pair.Seed(), issuer_pair.Seed(), recipient_pair.Address(), issuer_pair.Address(), asset_name, "0.1")
	
	set_trust(true, issue_pair.Seed(), issuer_pair.Seed(), "USD", issuer_pair.Address())
	//set_trust(false, issue_pair.Seed(), issuer_pair.Seed(), "USD", issuer_pair.Address())
	test_tx(issuer_pair.Seed(), issuer_pair.Seed(), issue_pair.Address(), issuer_pair.Address(), "USD", "0.1")
	test_tx(issue_pair.Seed(), issuer_pair.Seed(), issuer_pair.Address(), issuer_pair.Address(), "USD", "0.05")

	//test_tx(recipient_pair.Seed(), recipient_pair.Seed(), issuer_pair.Address(), asset_name, "0.1")
		
	
	issue_acct := LoadAccount(issue_pair.Address())	
	recipient_acct := LoadAccount(recipient_pair.Address())	
	
	PrintAccounts([]NamedAccount{issuer_acct, issue_acct, recipient_acct})
}

func set_trust(allow_trust bool, acct_seed string, signer string, asset_name string, issuer_acct string) {

	if (allow_trust) {
		log.Println("Adding trust for " + asset_name + " from " + issuer_acct + " to " + acct_seed)
		tx := b.Transaction(
			b.SourceAccount{acct_seed},
			b.AutoSequence{ClientEnv()},
			b.TestNetwork,
			b.Trust(
				asset_name,
				issuer_acct,
				//b.Limit("100.25"),
			),
		)
		txe := tx.Sign(signer)
		txeB64, err := txe.Base64()

		if err != nil {
			log.Fatal(err)
		}

		SubmitTxn(txeB64)
	} else {
		log.Println("Removing trust for " + asset_name + " from " + issuer_acct)
		tx := b.Transaction(
			b.SourceAccount{acct_seed},
			b.AutoSequence{ClientEnv()},
			b.TestNetwork,
			b.RemoveTrust(
				asset_name,
				issuer_acct,
			),
		)
		txe := tx.Sign(signer)
		txeB64, err := txe.Base64()

		if err != nil {
			log.Fatal(err)
		}

		SubmitTxn(txeB64)
	}
}

func test_tx(from string, signer string, to string, issuer string, asset_name string, amount string) {
    // // address: GB6S3XHQVL6ZBAF6FIK62OCK3XTUI4L5Z5YUVYNBZUXZ4AZMVBQZNSAU
    //from := "SCRUYGFG76UPX3EIUWGPIQPQDPD24XPR3RII5BD53DYPKZJGG43FL5HI"

    // // seed: SDLJZXOSOMKPWAK4OCWNNVOYUEYEESPGCWK53PT7QMG4J4KGDAUIL5LG
    //to := "GA3A7AD7ZR4PIYW6A52SP6IK7UISESICPMMZVJGNUTVIZ5OUYOPBTK6X"

    log.Println("Creating Txn: " + amount + " of " + asset_name + " from " + from + " to " + to + " signed by " + signer)
    tx := b.Transaction(
        b.SourceAccount{from},
        b.AutoSequence{ClientEnv()},
		b.TestNetwork,
        b.Payment(
            b.Destination{to},
            //b.NativeAmount{amount},
			b.CreditAmount{asset_name, issuer, amount},
        ),
    )

    log.Println("Signing Txn")
    txe := tx.Sign(signer)
    log.Println("Encoding Txn")
    txeB64, err := txe.Base64()
	//log.Println(txeB64)
	
    if err != nil {
        log.Fatal(err)
    }
	
	SubmitTxn(txeB64)
}

func add_signer(acct_seed string, signer string, change_signer string, weight uint32) {
	log.Printf("Adding weight %d for signer %s to %s", weight, change_signer, acct_seed)
	log.Println()
	tx := b.Transaction(
		b.SourceAccount{acct_seed},
        b.AutoSequence{ClientEnv()},
		b.TestNetwork,
		b.AddSigner(change_signer, weight),
		//b.MasterWeight(0),
		//b.InflationDest("GCT7S5BA6ZC7SV7GGEMEYJTWOBYTBOA7SC4JEYP7IAEDG7HQNIWKRJ4G"),
		//b.SetAuthRequired(),
		//b.SetAuthRevocable(),
		//b.SetAuthImmutable(),
		//b.ClearAuthRequired(),
		//b.ClearAuthRevocable(),
		//b.ClearAuthImmutable(),
		//b.SetThresholds(2, 3, 4),
		//b.HomeDomain("stellar.org"),
		//b.RemoveSigner(remove_signer),
	)

	txe := tx.Sign(acct_seed)
	txeB64, _ := txe.Base64()
	
	SubmitTxn(txeB64)
}

func master_weight(acct_seed string, signer string, weight uint32) {
	log.Println("Reweighting Master Key to " + fmt.Sprint(weight))
	tx := b.Transaction(
		b.SourceAccount{acct_seed},
        b.AutoSequence{ClientEnv()},
		b.TestNetwork,
		b.MasterWeight(weight),
		//b.AddSigner(change_signer, weight),
		//b.RemoveSigner(remove_signer),
		//b.InflationDest("GCT7S5BA6ZC7SV7GGEMEYJTWOBYTBOA7SC4JEYP7IAEDG7HQNIWKRJ4G"),
		//b.SetAuthRequired(),
		//b.SetAuthRevocable(),
		//b.SetAuthImmutable(),
		//b.ClearAuthRequired(),
		//b.ClearAuthRevocable(),
		//b.ClearAuthImmutable(),
		//b.SetThresholds(2, 3, 4),
		//b.HomeDomain("stellar.org"),
	)

	txe := tx.Sign(acct_seed)
	txeB64, _ := txe.Base64()
	
	SubmitTxn(txeB64)
}

func remove_signer(acct_seed string, signer string, remove_signer string) {
	tx := b.Transaction(
		b.SourceAccount{acct_seed},
        b.AutoSequence{ClientEnv()},
		b.TestNetwork,
		b.RemoveSigner(remove_signer),
		//b.ChangeSigner(change_signer, weight),
		//b.MasterWeight(0),
		//b.InflationDest("GCT7S5BA6ZC7SV7GGEMEYJTWOBYTBOA7SC4JEYP7IAEDG7HQNIWKRJ4G"),
		//b.SetAuthRequired(),
		//b.SetAuthRevocable(),
		//b.SetAuthImmutable(),
		//b.ClearAuthRequired(),
		//b.ClearAuthRevocable(),
		//b.ClearAuthImmutable(),
		//b.SetThresholds(2, 3, 4),
		//b.HomeDomain("stellar.org"),
	)

	txe := tx.Sign(signer)
	txeB64, _ := txe.Base64()
	
	SubmitTxn(txeB64)
}

func SubmitTxn(txeB64 string) {

    log.Println("Submitting Txn to Server: ", txeB64)

    resp, err := ClientEnv().SubmitTransaction(txeB64)
    if err != nil {
        log.Fatal(err)
    }

    log.Println("transaction posted in ledger:", resp.Ledger)
}

func MakePair(secret string) *keypair.Full {
    log.Println("Making Keypair for: " + secret)

	// MAKE THE KEYPAIR
	pair, err := keypair.FromRawSeed(hash.Hash([]byte(secret)))

    if err != nil {
        log.Fatal(err)
    }

    log.Println(pair.Seed())
    // SAV76USXIJOBMEQXPANUOQM6F5LIOTLPDIDVRJBFFE2MDJXG24TAPUU7
    log.Println(pair.Address())
    // GCFXHS4GXL6BVUCXBWXGTITROWLVYXQKQLF4YH5O5JT3YZXCYPAFBJZB
	
	return pair
}

func LoadAccount(address string) NamedAccount {
    log.Println("Loading Account Object: " + address)
	// LOAD ACCOUNT
	account, err := ClientEnv().LoadAccount(address)
	
    if err != nil {
        log.Fatal(err)
    }
	
	var named_account = NamedAccount{address, account}
	return named_account
}

func FundAccount(pair *keypair.Full) {
    address := pair.Address()
	log.Println("Create Account: " + address)

	// ASK FRIENDBOT TO FUND THE ACCOUNT
    resp, err := http.Get("https://horizon-testnet.stellar.org/friendbot?addr=" + address)
    if err != nil {
        log.Fatal(err)
    }

    defer resp.Body.Close()
    body, err := ioutil.ReadAll(resp.Body)
    if err != nil {
        log.Fatal(err)
    }
	//log.Println(body)
    fmt.Println(string((body)))

}

func PrintAccounts(accts []NamedAccount) {

	for _, acct := range accts {
		log.Println("Balances for account:", acct.Address)
		for _, balance := range acct.Account.Balances {
			log.Println(balance)
		}

		log.Println("Signers for account:", acct.Address)
		for _, signer := range acct.Account.Signers {
			log.Println(signer)
		}
	}
}
