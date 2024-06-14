package repository

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"

	"github.com/satyamvatstyagi/LinkManagementService/pkg/app/models"
	"github.com/satyamvatstyagi/LinkManagementService/pkg/common/cerr"
	"github.com/satyamvatstyagi/LinkManagementService/pkg/common/consts"
	"github.com/satyamvatstyagi/LinkManagementService/pkg/common/mtnapm"
	"go.elastic.co/apm/v2"
)

type userRepository struct {
	database *gorm.DB
}

func NewUserRepository(database *gorm.DB) models.LinkRepository {
	return &userRepository{
		database: database,
	}
}

func (u *userRepository) RegisterUser(ctx context.Context, userName string, password string) (string, error) {

	localUTCTime := time.Now()
	user := &models.User{
		UserName:  userName,
		Password:  password,
		CreatedAt: localUTCTime,
		UpdatedAt: localUTCTime,
	}

	//for fetching the database query
	statement := u.database.ToSQL(func(tx *gorm.DB) *gorm.DB {
		return tx.Create(user)
	})

	instrument := mtnapm.InitGormAPM(ctx, "postgresql", statement)
	defer instrument.GetSpan().End()

	if err := u.database.Create(user).Error; err != nil {
		// Check if err is of type *pgconn.PgError and error code is 23505, which is the error code for unique_violation
		if pgErr, ok := err.(*pgconn.PgError); ok && pgErr.Code == consts.UniqueViolation {
			apm.CaptureError(ctx, fmt.Errorf("db error: %s", pgErr.Error())).Send()
			log.Println("[LinkRepository][RegisterUser] User already exists for this user: ", pgErr.Error())
			return "", cerr.NewCustomErrorWithCodeAndOrigin("User already exists for this user", cerr.InvalidRequestErrorCode, err)
		}
		apm.CaptureError(ctx, fmt.Errorf("db error: %s", err.Error())).Send()
		log.Println("[LinkRepository][RegisterUser] Error in creating user: ", err)
		return "", err
	}

	return user.UUID.String(), nil
}

func (u *userRepository) GetUserByUserName(ctx context.Context, userName string) (*models.User, error) {
	var user models.User

	//for fetching the database query
	statement := u.database.ToSQL(func(tx *gorm.DB) *gorm.DB {
		return tx.Where("user_name = ?", userName).First(&user)
	})

	instrument := mtnapm.InitGormAPM(ctx, "postgresql", statement)
	defer instrument.GetSpan().End()

	if err := u.database.Where("user_name = ?", userName).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			apm.CaptureError(ctx, fmt.Errorf("db error: %s", err.Error())).Send()
			log.Println("[LinkRepository][GetUserByUserName] User not found: ", err)
			return nil, cerr.NewCustomErrorWithCodeAndOrigin("User not found", cerr.InvalidRequestErrorCode, err)
		}
		apm.CaptureError(ctx, fmt.Errorf("db error: %s", err.Error())).Send()
		log.Println("[LinkRepository][GetUserByUserName] Error in fetching user: ", err)
		return nil, err
	}
	return &user, nil
}

// Funcrtion to return list of string
func (u *userRepository) GetLinks(ctx context.Context) ([]string, error) {
	var list []string
	list = append(list, "https://t.me/hamSter_kombat_bot/start?startapp=kentId5706958952")
	list = append(list, "https://t.me/tronixapp_bot?start=5706958952")
	list = append(list, "https://t.me/mineralzbot?start=5706958952")
	list = append(list, "https://t.me/Demos_coin_bot/app?startapp=5706958952")
	list = append(list, "https://t.me/pixie_project_bot?start=5706958952")
	list = append(list, "https://t.me/pixelversexyzbot?start=5706958952")
	return list, nil
}
