package approve

import (
	"context"
	"mime/multipart"

	"github.com/JIeeiroSst/workflow-service/common"
	"github.com/JIeeiroSst/workflow-service/dto"
	"github.com/JIeeiroSst/workflow-service/internal/activities/approve/facade"
	"github.com/JIeeiroSst/workflow-service/pkg/excel"
	"github.com/JIeeiroSst/workflow-service/pkg/snowflake"
)

type ProcessState struct {
	Facade      facade.Facade
	GoogleSheet excel.GoogleSheet
	FileHeader  *multipart.FileHeader
}

type Upload struct {
	Type          string
	File          string
	ActiveUser    ActiveUser
	SpreadsheetId string
	Sheet         string
	FilterStruct  interface{}
	Status        Status
}

type Process struct {
	Type       string
	Email      string
	IsApprove  bool
	ActiveUser ActiveUser
	RecordId   int
	Status     Status
	BatchID    string
}

type Approve struct {
	Type       string
	Email      string
	IsApprove  bool
	ActiveUser ActiveUser
	RecordId   int
	Status     Status
	BatchID    string
}

type NotiSlack struct {
	IsApprove bool
	Message   string
}

func (p *ProcessState) UploadApprove(upload Upload) {
	id := snowflake.GearedID()
	switch upload.Type {
	case common.Local:
		sheetFile, _ := excel.ReadFileExcel(p.FileHeader)
		sheets := sheetFile.Sheet
		sheet := sheets[common.SheetLocal]
		switch upload.FilterStruct.(type) {
		case dto.BestSellingPlayStationRequestDTO:
			autoGenerateds := excel.GetRowValues(sheet)
			sellingPlayStation := dto.FormatLocalSellingPlayStation(autoGenerateds)
			p.Facade.SellingPlayStation.InsertSellingPlayStation(sellingPlayStation, id)
		case dto.GameRequestDTO:
			autoGenerateds := excel.GetRowValues(sheet)
			games := dto.FormatLocalGame(autoGenerateds)
			p.Facade.Game.InsertGame(games, id)
		case dto.SeattleWeatherRequestDTO:
			autoGenerateds := excel.GetRowValues(sheet)
			seattleWeather := dto.FormatLocalSeattleWeather(autoGenerateds)
			p.Facade.SeattleWeather.InsertSeattleWeather(seattleWeather, id)
		case dto.SpotifyQuarterlyRequestDTO:
			autoGenerateds := excel.GetRowValues(sheet)
			spotifyQuarterly := dto.FormatLocalSpotifyQuarterly(autoGenerateds)
			p.Facade.SpotifyQuarterly.InsertSpotifyQuarterly(spotifyQuarterly, id)
		}
	case common.GoogleSheet:
		switch upload.FilterStruct.(type) {
		case dto.BestSellingPlayStationRequestDTO:
			autoGenerateds, err := p.GoogleSheet.SheetGoogleApi(upload.SpreadsheetId, upload.Sheet)
			if err != nil {
				return
			}
			sellingPlayStation := dto.FormatBestSellingPlayStation(autoGenerateds)
			p.Facade.SellingPlayStation.InsertSellingPlayStation(sellingPlayStation, id)
		case dto.GameRequestDTO:
			autoGenerateds, err := p.GoogleSheet.SheetGoogleApi(upload.SpreadsheetId, upload.Sheet)
			if err != nil {
				return
			}
			games := dto.FormatGame(autoGenerateds)
			p.Facade.Game.InsertGame(games, id)
		case dto.SeattleWeatherRequestDTO:
			autoGenerateds, err := p.GoogleSheet.SheetGoogleApi(upload.SpreadsheetId, upload.Sheet)
			if err != nil {
				return
			}
			seattleWeathes := dto.FormatSeattleWeather(autoGenerateds)
			p.Facade.SeattleWeather.InsertSeattleWeather(seattleWeathes, id)
		case dto.SpotifyQuarterlyRequestDTO:
			autoGenerateds, err := p.GoogleSheet.SheetGoogleApi(upload.SpreadsheetId, upload.Sheet)
			if err != nil {
				return
			}
			spotifyQuarterlies := dto.FormatSpotifyQuarterly(autoGenerateds)
			p.Facade.SpotifyQuarterly.InsertSpotifyQuarterly(spotifyQuarterlies, id)
		}
	}
	upload.ActiveUser.Status = PENDING
	activeUser := FormatActiveUser(upload.ActiveUser)
	if err := p.Facade.ActiveUser.InsertActiveUser(activeUser, id); err != nil {
		return
	}
}

func (p *ProcessState) ProcessApprove(process Process) {
	process.ActiveUser.Status = PROCESS
	activeUser := FormatActiveUser(process.ActiveUser)
	if err := p.Facade.ActiveUser.UpdateActiveUser(process.BatchID, activeUser); err != nil {
		return
	}
}

func (p *ProcessState) Approve(approve Approve) {
	approve.ActiveUser.Status = APPROVE
	if !approve.IsApprove {
		approve.ActiveUser.Status = REJECT
	}
	activeUser := FormatActiveUser(approve.ActiveUser)
	if err := p.Facade.ActiveUser.UpdateActiveUser(approve.BatchID, activeUser); err != nil {
		return
	}
}

func (a *ProcessState) ApproveProcess(_ context.Context, process ProcessState) error {
	return nil
}

func (a *ProcessState) SendAbandonedProcess(_ context.Context, noti NotiSlack) error {

	return nil
}
