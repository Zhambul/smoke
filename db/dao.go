package db

import (
	"smoke3/domain"
	"errors"
	"log"
	"strings"
)

var NotFound = errors.New("Not Found")
var NotUnique = errors.New("Not Unique")

func ChangeGroupName(g *domain.Group) error {
	_, err := db.Exec(`UPDATE "group" SET name=$1 WHERE id=$2`, g.Name, g.Id)
	return err
}

func CreateNewGroup(account *domain.Account, groupName string) (*domain.Group, error) {
	log.Println("createNewGroup START")
	tx, _ := db.Begin()
	row := db.QueryRow(`INSERT INTO "group" (name, creator_account_id, uuid)
 							VALUES ($1, $2, (SELECT uuid_generate_v4()))
 							RETURNING id, uuid`, groupName, account.Id)

	var groupId int
	var uuid string
	err := row.Scan(&groupId, &uuid)
	if err != nil {
		tx.Rollback()
		if strings.HasPrefix(err.Error(), "pq: duplicate key value violates unique constraint") {
			return nil, NotUnique
		}
		return nil, err
	}
	if account.Id == 0 {
		return nil, errors.New("account has no id")
	}
	db.Exec("INSERT INTO group_account(account_id,group_id) VALUES ($1, $2)", account.Id, groupId)

	tx.Commit()

	log.Println("createNewGroup END")
	return &domain.Group{
		Id:             groupId,
		Name:           groupName,
		UUID:           uuid,
		CreatorAccount: account,
	}, nil
}

func GetGroupsByAccount(account *domain.Account) ([]*domain.Group, error) {
	log.Println("getCompaniesByAccount START")

	rows, err := db.Query(`SELECT c.* FROM "group" c
	INNER JOIN group_account ca
	ON c.id = ca.group_id
	AND ca.account_id = ($1)`, account.Id)

	if err != nil {
		return nil, err
	}

	groups := make([]*domain.Group, 0)

	for rows.Next() {
		group := &domain.Group{}
		var creatorAccountId int
		rows.Scan(&group.Id, &group.Name, &creatorAccountId, &group.UUID)
		log.Printf("creatorAccountId - %v\n", creatorAccountId)
		populateAccountsToGroup(group, creatorAccountId)
		groups = append(groups, group)
	}

	log.Println("getCompaniesByAccount END")

	return groups, nil
}

func CreateAccount(firstName string, lastName string, chatId int) (*domain.Account, error) {
	log.Println("createAccount START")
	row := db.QueryRow("INSERT INTO account(first_name, last_name, chat_id) VALUES ($1, $2, $3) RETURNING id", firstName, lastName, chatId)
	var accountId int
	row.Scan(&accountId)
	log.Println("createAccount END")
	return &domain.Account{
		Id:        accountId,
		FirstName: firstName,
		LastName:  lastName,
		ChatId:    chatId,
	}, nil
}

func GetAccountByChatId(chatId int) (*domain.Account, error) {
	log.Println("getAccountByChatId START")
	acc := &domain.Account{}
	err := db.QueryRow("SELECT * FROM account WHERE chat_id = $1", chatId).Scan(&acc.Id, &acc.FirstName, &acc.LastName, &acc.ChatId)
	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			return nil, NotFound
		}
		return nil, err
	}
	log.Println("getAccountByChatId END")
	return acc, nil
}

func GetGroupByUUID(uuid string) (*domain.Group, error) {
	log.Println("GetGroupByName START")
	group := domain.Group{}
	var creatorAccountId int
	log.Println("SELECT * FROM group WHERE uuid=" + uuid)
	err := db.QueryRow("SELECT * FROM \"group\" WHERE uuid=$1", uuid).Scan(&group.Id, &group.Name, &creatorAccountId, &group.UUID)
	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			//panic(err)
			return nil, NotFound
		}
		return nil, err
	}
	if err := populateAccountsToGroup(&group, creatorAccountId); err != nil {
		return nil, err
	}
	log.Println("GetGroupByName END")
	return &group, nil
}

func AddAccountToGroup(account *domain.Account, group *domain.Group) error {
	log.Println("AddAccountToGroup START")
	_, err := db.Exec("INSERT INTO group_account(account_id, group_id) VALUES($1, $2)", account.Id, group.Id)
	if err != nil {
		if err.Error() == "pq: duplicate key value violates unique constraint \"uq_group_account\"" {
			return NotUnique
		}
	}
	group.Accounts = append(group.Accounts, account)
	log.Println("AddAccountToGroup END")
	return err
}

func LeaveGroup(group *domain.Group, acc *domain.Account) error {
	_, err := db.Exec("DELETE FROM group_account WHERE account_id = $1 AND group_id = $2", acc.Id, group.Id)
	return err
}

func DeleteGroup(group *domain.Group) error {
	_, err := db.Exec("DELETE FROM \"group\" WHERE id = $1", group.Id)
	return err
}

func populateAccountsToGroup(group *domain.Group, creatorAccountId int) error {
	log.Println("populateAccountsToGroup START")
	accountsRow, err := db.Query(`SELECT a.* FROM account a
		INNER JOIN group_account ca
		ON a.id = ca.account_id AND
		ca.group_id = ($1)`, group.Id)

	if err != nil {
		return err
	}

	for accountsRow.Next() {
		account := &domain.Account{}
		accountsRow.Scan(&account.Id, &account.FirstName, &account.LastName, &account.ChatId)
		group.Accounts = append(group.Accounts, account)
	}

	for _, acc := range group.Accounts {
		log.Printf("account %v in group", acc)
		if acc.Id == creatorAccountId {
			group.CreatorAccount = acc
		}
	}
	log.Println("populateAccountsToGroup END")
	return nil
}
