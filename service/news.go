package service

import (
	"app/models"
	"errors"
	"github.com/beego/beego/v2/client/orm"
	"time"
)

func News(creatorID int64) ([]models.News, error) {
	o := getReadOrm()
	var ids []int64
	var text []string
	var creatorIDs []int64
	var times []time.Time
	_, err := o.Raw(
		`SELECT id, text, creator_id, create_timestamp FROM news WHERE creator_id=? ORDER BY create_timestamp DESC;`,
		creatorID,
	).QueryRows(&ids, &text, &creatorIDs, &times)
	if err != nil {
		if errors.Is(err, orm.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	news := make([]models.News, 0, len(ids))
	for i, id := range ids {
		news = append(news, models.News{ID: id, Text: text[i], Timestamp: times[i], Creator: models.User{Id: creatorID}})
	}

	return news, nil
}

func AddNews(creatorID int64, text string) error {
	t := time.Now().UTC()
	res, err := orm.NewOrm().Raw("INSERT INTO news(create_timestamp, creator_id, text) VALUES(?, ?, ?) ", t, creatorID, text).Exec()
	if err != nil {
		return err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return err
	}

	SendNewsMessage(models.News{ID: id, Text: text, Creator: models.User{Id: creatorID}, Timestamp: t})
	SendRabbitNewsMessage(creatorID, models.News{ID: id, Text: text, Creator: models.User{Id: creatorID}, Timestamp: t})
	return nil
}

func GetFriendsNews(userID int64) ([]models.News, error) {
	o := getReadOrm()

	var ids []int64
	var text []string
	var creatorIDs []int64
	var times []time.Time
	_, err := o.Raw(`
SELECT n.id, text, creator_id, create_timestamp 
FROM news n
JOIN (
	SELECT user_id_1 AS id FROM friend f WHERE user_id_2 = ?
	UNION
	SELECT user_id_2 AS id FROM friend f WHERE user_id_1 = ?
) f ON n.creator_id = f.id
ORDER BY create_timestamp DESC
LIMIT 100;`,
		userID, userID,
	).QueryRows(&ids, &text, &creatorIDs, &times)
	if err != nil {
		if errors.Is(err, orm.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	news := make([]models.News, 0, len(ids))
	for i, id := range ids {
		news = append(news, models.News{ID: id, Text: text[i], Timestamp: times[i], Creator: models.User{Id: creatorIDs[i]}})
	}

	return news, nil
}

func AddFriendsNews(news models.News) error {
	o := getReadOrm()
	var friendsIDs []int64
	_, err := o.Raw(`
SELECT user_id_1 AS id FROM friend f WHERE user_id_2 = ?
UNION
SELECT user_id_2 AS id FROM friend f WHERE user_id_1 = ?
`, news.Creator.Id, news.Creator.Id).QueryRows(&friendsIDs)
	if err != nil {
		return err
	}

	for _, id := range friendsIDs {
		if IsActiveUser(id) {
			AddCacheNews(id, []models.News{news})
		} else {
			RemoveCacheNews(id)
		}
	}
	return nil
}
