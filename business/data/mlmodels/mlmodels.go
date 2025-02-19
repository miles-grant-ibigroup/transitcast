package mlmodels

import (
	"fmt"
	"github.com/jmoiron/sqlx"
	"time"
)

// MLModelType stores the types of models the system currently knows how to train
type MLModelType struct {
	MLModelTypeId int `db:"ml_model_type_id" json:"ml_model_type_id"`
	Name          string
}

// MLModel stores definitions for each model trained or to be trained by the system
type MLModel struct {
	MLModelId                    int64          `db:"ml_model_id" json:"ml_model_id"`
	Version                      int            `db:"version" json:"version"`
	StartTimestamp               time.Time      `db:"start_timestamp" json:"start_timestamp"`
	EndTimestamp                 time.Time      `db:"end_timestamp" json:"end_timestamp"`
	MLModelTypeId                int            `db:"ml_model_type_id" json:"ml_model_type_id"`
	TrainFlag                    bool           `db:"train_flag" json:"train_flag"`
	TrainedTimestamp             *time.Time     `db:"trained_timestamp" json:"trained_timestamp"`
	AvgRMSE                      float64        `db:"avg_rmse" json:"avg_rmse"`
	MLRMSE                       float64        `db:"ml_rmse" json:"ml_rmse"`
	FeatureTrainedStartTimestamp *time.Time     `db:"feature_trained_start_timestamp" json:"feature_trained_start_timestamp"`
	FeatureTrainedEndTimestamp   *time.Time     `db:"feature_trained_end_timestamp" json:"feature_trained_end_timestamp"`
	ModelName                    string         `db:"model_name" json:"model_name"`
	CurrentlyRelevant            bool           `db:"currently_relevant" json:"currently_relevant"`
	LastTrainAttemptTimestamp    *time.Time     `db:"last_train_attempt_timestamp" json:"last_train_attempt_timestamp"`
	ModelStops                   []*MLModelStop `json:"model_stops"`
}

// MLModelStop defines stops included in each model
type MLModelStop struct {
	MLModelStopId int64  `db:"ml_model_stop_id" json:"ml_model_stop_id"`
	MLModelId     int64  `db:"ml_model_id" json:"ml_model_id"`
	Sequence      int    `db:"sequence" json:"sequence"`
	StopId        string `db:"stop_id" json:"stop_id"`
	NextStopId    string `db:"next_stop_id" json:"next_stop_id"`
}

//GetMLModelType loads MLModelType with ml_model_type of modelTypeName
func GetMLModelType(db *sqlx.DB, modelTypeName string) (*MLModelType, error) {
	query := "select * from ml_model_type where name = $1"
	var modelType MLModelType
	err := db.Get(&modelType, query, modelTypeName)
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve ModelType %s. error: %w", modelTypeName, err)
	}
	return &modelType, nil
}

//MakeMLModel MLModelType factory
func MakeMLModel(modelType *MLModelType,
	version int,
	at time.Time,
	modelName string) *MLModel {

	return &MLModel{
		Version:           version,
		StartTimestamp:    at,
		EndTimestamp:      time.Date(9999, 12, 31, 23, 29, 59, 0, at.Location()),
		MLModelTypeId:     modelType.MLModelTypeId,
		TrainFlag:         true,
		CurrentlyRelevant: true,
		ModelName:         modelName,
		ModelStops:        make([]*MLModelStop, 0),
	}
}

//MakeMLModelStop MLModelStop factory
func MakeMLModelStop(sequence int, stopId string, nextStopId string) *MLModelStop {
	return &MLModelStop{
		Sequence:   sequence,
		StopId:     stopId,
		NextStopId: nextStopId,
	}
}

//RecordNewMLModel inserts new MLModel record
func RecordNewMLModel(db *sqlx.DB, model *MLModel) (*MLModel, error) {
	statementString := "insert into ml_model " +
		"(version, " +
		"start_timestamp, " +
		"end_timestamp, " +
		"ml_model_type_id, " +
		"model_name, " +
		"train_flag, " +
		"trained_timestamp, " +
		"avg_rmse, " +
		"ml_rmse, " +
		"feature_trained_start_timestamp, " +
		"feature_trained_end_timestamp," +
		"currently_relevant, " +
		"last_train_attempt_timestamp) " +
		"values (:version, " +
		":start_timestamp, " +
		":end_timestamp, " +
		":ml_model_type_id, " +
		":model_name, " +
		":train_flag, " +
		":trained_timestamp, " +
		":avg_rmse, " +
		":ml_rmse, " +
		":feature_trained_start_timestamp, " +
		":feature_trained_end_timestamp, " +
		":currently_relevant, " +
		":last_train_attempt_timestamp )"
	if model.MLModelId != 0 {
		statementString = "update ml_model set version = :version, " +
			"start_timestamp = :start_timestamp, " +
			"end_timestamp = :end_timestamp, " +
			"ml_model_type_id = :ml_model_type_id, " +
			"model_name = :model_name, " +
			"train_flag = :train_flag," +
			"trained_timestamp = :trained_timestamp, " +
			"avg_rmse = :avg_rmse, " +
			"ml_rmse = :ml_rmse, " +
			"feature_trained_start_timestamp = :feature_trained_start_timestamp, " +
			"feature_trained_end_timestamp = :feature_trained_end_timestamp, " +
			"currently_relevant = :currently_relevant, " +
			"last_train_attempt_timestamp = :last_train_attempt_timestamp " +
			"where ml_model_id = :ml_model_id"
	}
	statementString = db.Rebind(statementString)
	_, err := db.NamedExec(statementString, model)
	if err != nil {
		return nil, err
	}

	statementString = db.Rebind("select ml_model_id from ml_model " +
		"where model_name = ? " +
		"and start_timestamp = ? " +
		"and end_timestamp = ? limit 1")
	err = db.Get(&model.MLModelId, statementString, model.ModelName, model.StartTimestamp, model.EndTimestamp)
	if err != nil {
		return nil, err
	}

	for _, modelStop := range model.ModelStops {
		modelStop.MLModelId = model.MLModelId
		_, err := RecordNewMLStopModel(db, modelStop)
		if err != nil {
			return nil, err
		}
	}
	return model, nil
}

//UpdateMLModel updates existing MLModel record
func UpdateMLModel(db *sqlx.DB, model *MLModel) (*MLModel, error) {
	statementString := "update ml_model set version = :version, " +
		"start_timestamp = :start_timestamp, " +
		"end_timestamp = :end_timestamp, " +
		"ml_model_type_id = :ml_model_type_id, " +
		"model_name = :model_name, " +
		"train_flag = :train_flag," +
		"trained_timestamp = :trained_timestamp, " +
		"avg_rmse = :avg_rmse, " +
		"ml_rmse = :ml_rmse, " +
		"feature_trained_start_timestamp = :feature_trained_start_timestamp, " +
		"feature_trained_end_timestamp = :feature_trained_end_timestamp, " +
		"currently_relevant = :currently_relevant, " +
		"last_train_attempt_timestamp = :last_train_attempt_timestamp " +
		"where ml_model_id = :ml_model_id"
	statementString = db.Rebind(statementString)
	_, err := db.NamedExec(statementString, model)
	if err != nil {
		return nil, err
	}
	return model, nil
}

//RecordNewMLStopModel records new MLModelStop record.
func RecordNewMLStopModel(db *sqlx.DB, modelStop *MLModelStop) (*MLModelStop, error) {

	statementString := "insert into ml_model_stop (ml_model_id, sequence, stop_id, next_stop_id) " +
		"values (:ml_model_id, :sequence, :stop_id, :next_stop_id)"
	statementString = db.Rebind(statementString)
	_, err := db.NamedExec(statementString, modelStop)
	if err != nil {
		return nil, err
	}
	statementString = db.Rebind("select ml_model_stop_id from ml_model_stop " +
		"where ml_model_id = ? and sequence = ?")
	err = db.Get(&modelStop.MLModelStopId, statementString, modelStop.MLModelId, modelStop.Sequence)
	if err != nil {
		return nil, err
	}
	return modelStop, nil
}

//GetAllCurrentMLModelsByName returns map of all current MLModel by ModelName, where current timestamp is between
//ml_model.start_timestamp and ml_model.end_timestamp
func GetAllCurrentMLModelsByName(db *sqlx.DB) (map[string]*MLModel, error) {
	modelStopMap, err := getAllCurrentMLModelStopsByMLModelID(db)
	if err != nil {
		return nil, err
	}

	statementString := "select version, " +
		"start_timestamp, " +
		"end_timestamp, " +
		"ml_model_type_id, " +
		"model_name, " +
		"train_flag, " +
		"trained_timestamp, " +
		"avg_rmse, " +
		"ml_rmse, " +
		"feature_trained_start_timestamp, " +
		"feature_trained_end_timestamp," +
		"currently_relevant, " +
		"last_train_attempt_timestamp " +
		"from ml_model where current_timestamp between start_timestamp and end_timestamp"
	rows, err := db.Queryx(statementString)
	if err != nil {
		return nil, err
	}
	modelMap := make(map[string]*MLModel)
	for rows.Next() {
		model := MLModel{}
		err = rows.StructScan(&model)
		if err != nil {
			return nil, err
		}
		model.ModelStops = modelStopMap[model.MLModelId]
		modelMap[model.ModelName] = &model
	}
	return modelMap, nil
}

func getAllCurrentMLModelStopsByMLModelID(db *sqlx.DB) (map[int64][]*MLModelStop, error) {
	statementString := "select s.ml_model_id, s.ml_model_stop_id, s.stop_id, s.next_stop_id, s.sequence " +
		"from ml_model_stop s left join ml_model m on s.ml_model_id = m.ml_model_id " +
		"where current_timestamp between m.start_timestamp and m.end_timestamp " +
		"order by s.ml_model_id, s.sequence"
	rows, err := db.Queryx(statementString)
	if err != nil {
		return nil, err
	}
	stopMap := make(map[int64][]*MLModelStop)
	for rows.Next() {
		stop := MLModelStop{}
		err = rows.StructScan(&stop)
		if err != nil {
			return nil, err
		}
		modelStops, present := stopMap[stop.MLModelId]
		if !present {
			modelStops = make([]*MLModelStop, 1)
			stopMap[stop.MLModelId] = modelStops
		}
		modelStops = append(modelStops, &stop)
	}
	return stopMap, nil
}
