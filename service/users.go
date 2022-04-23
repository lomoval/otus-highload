package service

import (
	"app/models"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"github.com/beego/beego/v2/client/orm"
	"github.com/go-sql-driver/mysql"
	"math/rand"
	"strconv"
	"time"
)

var ErrDuplicate = errors.New("duplicate value in column")

const (
	maxAgeYears = 150
	minAgeYears = 8

	sqlErrCodeDuplicate = 1062
)

var SlavesCount = 0

func getReadOrm() orm.Ormer {
	if SlavesCount == 0 {
		return orm.NewOrm()
	}
	return orm.NewOrmUsingDB("slave" + strconv.Itoa(rand.Intn(SlavesCount)))
}

func HashPassword(password string) string {
	return hex.EncodeToString(sha256.New().Sum([]byte(password)))
}

func CreateUser(user models.User, password string) (models.User, error) {
	o := orm.NewOrm()
	err := o.DoTx(func(ctx context.Context, txOrm orm.TxOrmer) error {
		res, err := o.Raw(`INSERT INTO user (login, password) VALUES(?,?);`, user.Login, HashPassword(password)).Exec()
		if err != nil {
			return err
		}
		user.Id, err = res.LastInsertId()
		if err != nil {
			return err
		}
		res, err = o.Raw(
			`INSERT INTO profile (user_id, name, surname, birth_date, sex_id, city) VALUES(?,?,?,?,?,?);`,
			user.Id,
			user.Profile.Name,
			user.Profile.Surname,
			user.Profile.BirthDate,
			user.Profile.Sex.Id,
			user.Profile.City).Exec()

		if err != nil {
			return err
		}

		user.Profile.Id, err = res.LastInsertId()
		return err
	})

	if err != nil {
		var sqlErr *mysql.MySQLError
		if errors.As(err, &sqlErr) {
			if sqlErr.Number == sqlErrCodeDuplicate {
				return models.User{}, ErrDuplicate
			}
		}
		return models.User{}, err
	}

	return user, nil
}

func SaveUser(user models.User) error {
	o := orm.NewOrm()
	err := o.DoTx(func(ctx context.Context, txOrm orm.TxOrmer) error {
		_, err := o.Raw(
			`UPDATE profile SET name=?, surname=?, birth_date=?, sex_id=?, city=? WHERE id = ?;`,
			user.Profile.Name,
			user.Profile.Surname,
			user.Profile.BirthDate,
			user.Profile.Sex.Id,
			user.Profile.City,
			user.Profile.Id,
		).Exec()
		if err != nil {
			return err
		}

		_, err = o.Raw(`DELETE FROM interest WHERE user_id = ?`, user.Id).Exec()
		if err != nil {
			return err
		}

		for _, interest := range user.Interests {
			_, err = o.Raw(`INSERT INTO interest(user_id, name) VALUES(?,?)`, user.Id, interest.Name).Exec()
			if err != nil {
				return err
			}
		}
		return nil
	})
	return err
}

func GetUserLoginInfo(login string, password string) (*models.User, error) {
	o := orm.NewOrm()
	var user models.User

	err := o.Raw("SELECT u.id, login, p.Id, name "+
		"FROM user u join profile p ON u.id = p.user_id "+
		"WHERE login = ? AND password = ?", login, HashPassword(password)).
		QueryRow(&user.Id, &user.Login, &user.Profile.Id, &user.Profile.Name)

	if err != nil {
		if errors.Is(err, orm.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	if user.Id > 0 {
		return &user, nil
	}
	return nil, nil
}

func Friends(user models.User, limit int, offset int) ([]orm.Params, error) {
	o := orm.NewOrm()
	var maps []orm.Params
	_, err := o.Raw(`
SELECT u.id AS Id, p.name AS Name, p.surname AS Surname 
FROM user u
JOIN (
	SELECT user_id_1 AS id FROM friend f WHERE user_id_2 = ?
	UNION
	SELECT user_id_2 AS id FROM friend f WHERE user_id_1 = ?
) f ON u.id = f.id
JOIN profile p ON p.user_id = u.id
LIMIT ? OFFSET ?`, user.Id, user.Id, limit, offset).Values(&maps)
	if err != nil {
		return nil, err
	}
	return maps, nil
}

func Users(user models.User, limit int, offset int) ([]orm.Params, error) {
	o := orm.NewOrm()

	var maps []orm.Params
	_, err := o.Raw(`
SELECT u.id AS Id, p.name AS Name, p.surname AS Surname 
FROM user u
JOIN profile p ON p.user_id = u.id
WHERE u.id <> ?
AND u.id NOT IN (
SELECT user_id_1 AS id FROM friend f WHERE user_id_2 = ?
	UNION
	SELECT user_id_2 AS id FROM friend f WHERE user_id_1 = ?
)
LIMIT ? OFFSET ?`,
		user.Id, user.Id, user.Id, limit, offset).Values(&maps)

	if err != nil {
		return nil, err
	}
	return maps, nil
}

func FindUsers(user models.User, limit int, offset int, name string, surname string) ([]orm.Params, error) {
	o := getReadOrm()

	var maps []orm.Params
	_, err := o.Raw(`
SELECT u.id AS Id, p.name AS Name, p.surname AS Surname 
FROM user u
JOIN profile p ON p.user_id = u.id
WHERE u.id <> ?
AND p.Name LIKE ? AND p.Surname LIKE ? 
AND u.id NOT IN (
SELECT user_id_1 AS id FROM friend f WHERE user_id_2 = ?
	UNION
	SELECT user_id_2 AS id FROM friend f WHERE user_id_1 = ?
)
ORDER BY u.Id ASC LIMIT ? OFFSET ?`,
		user.Id, name+"%", surname+"%", user.Id, user.Id, limit, offset).Values(&maps)

	if err != nil {
		return nil, err
	}
	return maps, nil
}

func FindUsersByInterest(user models.User, limit int, offset int, interest string) ([]orm.Params, error) {
	o := getReadOrm()

	var maps []orm.Params
	_, err := o.Raw(`
SELECT u.id AS Id, p.name AS Name, p.surname AS Surname 
FROM user u
JOIN profile p ON p.user_id = u.id
JOIN interest i ON i.user_id = u.id
WHERE u.id <> ?
AND i.Name = ? 
AND u.id NOT IN (
	SELECT user_id_1 AS id FROM friend f WHERE user_id_2 = ?
	UNION
	SELECT user_id_2 AS id FROM friend f WHERE user_id_1 = ?
)
ORDER BY u.Id ASC LIMIT ? OFFSET ?`,
		user.Id, interest, user.Id, user.Id, limit, offset).Values(&maps)

	if err != nil {
		return nil, err
	}
	return maps, nil
}

func Profile(id int64) (models.User, error) {
	o := orm.NewOrm()

	var u models.User
	err := o.Raw(`
SELECT u.id, p.id, p.name, p.surname, p.birth_date AS Birthdate, s.Id, s.Name, p.City 
FROM user u
JOIN profile p ON p.user_id = u.id
JOIN sex s ON s.id = p.sex_id
WHERE u.id = ?`, id).QueryRow(
		&u.Id,
		&u.Profile.Id,
		&u.Profile.Name,
		&u.Profile.Surname,
		&u.Profile.BirthDate,
		&u.Profile.Sex.Id,
		&u.Profile.Sex.Name,
		&u.Profile.City,
	)

	if err != nil {
		return models.User{}, err
	}

	var ids []int
	var interests []string
	_, err = o.Raw(`
SELECT id, name 
FROM interest u
WHERE user_id = ?`, id).QueryRows(&ids, &interests)

	if err != nil {
		return models.User{}, err
	}

	u.Interests = make([]models.Interest, len(ids))
	for i, id := range ids {
		u.Interests[i] = models.Interest{Id: id, Name: interests[i]}
	}

	return u, nil
}

func AddFriend(userID int64, friendID int64) error {
	if userID == friendID {
		return nil
	}
	userID, friendID = friendsLinkIds(userID, friendID)

	_, err := orm.NewOrm().Raw("INSERT INTO friend(user_id_1, user_id_2) VALUES(?, ?) ", userID, friendID).Exec()
	return err
}

func RemoveFriend(userID int64, friendID int64) error {
	if userID == friendID {
		return nil
	}
	userID, friendID = friendsLinkIds(userID, friendID)
	_, err := orm.NewOrm().Raw("DELETE FROM friend WHERE user_id_1 = ? AND user_id_2 = ?", userID, friendID).Exec()
	return err
}

func ValidateProfileData(profile models.Profile) bool {
	return !(profile.Name == "" ||
		profile.Surname == "" ||
		profile.BirthDate.Year() < time.Now().UTC().Year()-maxAgeYears ||
		profile.BirthDate.After(truncateToDay(time.Now().UTC().AddDate(-minAgeYears, 0, 0))) ||
		profile.Sex.Id <= 0 ||
		profile.Sex.Id > 3)
}

func friendsLinkIds(userID int64, friendID int64) (int64, int64) {
	if userID > friendID {
		return friendID, userID
	}
	return userID, friendID
}

func truncateToDay(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
}
