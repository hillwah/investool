// 封装需要导出数据结果

package exportor

import (
	"context"
	"sort"
	"strings"

	"github.com/axiaoxin-com/goutils"
	"github.com/axiaoxin-com/x-stock/model"
)

// Data 数据模板
type Data struct {
	// 股票名
	Name string `json:"name"                      csv:"股票名"`
	// 股票代码
	Code string `json:"code"                      csv:"股票代码"`
	// 所属行业
	Industry string `json:"industry"                  csv:"所属行业"`
	// 题材关键词
	Keywords string `json:"keywords"                  csv:"题材关键词"`
	// 公司信息
	CompanyProfile string `json:"company_profile"           csv:"公司信息"`
	// 主营构成
	MainForms string `json:"main_forms"                csv:"主营构成"`
	// 本业营收比
	BYYSRatio float64 `json:"byys_ration"               csv:"本业营收比"`
	// 财报年份-类型
	ReportDateName string `json:"report_date_name"          csv:"数据源"`
	// 财报审计意见
	ReportOpinion interface{} `json:"report_opinion"            csv:"财报审计意见"`
	// 价值评估
	JZPG string `json:"jzpg"                      csv:"价值评估"`
	// 最新一期 ROE
	LatestROE float64 `json:"latest_roe"                csv:"最新一期 ROE"`
	// ROE 同比增长
	ROETBZZ float64 `json:"roe_tbzz"                  csv:"ROE 同比增长 (%)"`
	// 近五年 ROE
	ROE5Y []float64 `json:"roe_5y"                    csv:"近五年 ROE"`
	// 最新一期 EPS
	LatestEPS float64 `json:"latest_eps"                csv:"最新一期 EPS"`
	// EPS 同比增长
	EPSTBZZ float64 `json:"eps_tbzz"                  csv:"EPS 同比增长 (%)"`
	// 近五年 EPS
	EPS5Y []float64 `json:"eps_5y"                    csv:"近五年 EPS"`
	// 营业总收入
	TotalIncome interface{} `json:"total_income"              csv:"营业总收入"`
	// 营业总收入同比增长
	TotalIncomeTBZZ float64 `json:"total_income_tbzz"         csv:"营业总收入同比增长 (%)"`
	// 近五年营收
	TotalIncome5Y []float64 `json:"total_income_5y"           csv:"近五年营收"`
	// 归属净利润
	NetProfit interface{} `json:"net_profit"                csv:"归属净利润（元）"`
	// 归属净利润同比增长 (%)
	NetProfitTBZZ float64 `json:"net_profit_tbzz"           csv:"归属净利润同比增长 (%)"`
	// 近五年净利润
	NetProfit5Y []float64 `json:"net_profit_5y"             csv:"近五年净利润"`
	// 最新股息率 (%)
	ZXGXL float64 `json:"zxgxl"                     csv:"最新股息率 (%)"`
	// 财报披露日期
	FinaReportDate string `json:"fina_report_date"          csv:"财报披露日期"`
	// 预约财报披露日期
	FinaAppointPublishDate string `json:"fina_appoint_publish_date" csv:"预约财报披露日期"`
	// 实际财报披露日期
	FinaActualPublishDate string `json:"fina_actual_publish_date"  csv:"实际财报披露日期"`
	// 总市值（字符串）
	TotalMarketCap interface{} `json:"total_market_cap"          csv:"总市值"`
	// 当时价格
	Price float64 `json:"price"                     csv:"价格"`
	// 估算合理价格
	RightPrice interface{} `json:"right_price"               csv:"估算合理价格"`
	// 合理价格与当时价的价格差
	PriceSpace interface{} `json:"price_space"               csv:"合理价差"`
	// 历史波动率 (%)
	HV float64 `json:"hv"                        csv:"历史波动率 (%)"`
	// 最新负债率 (%)
	ZXFZL float64 `json:"zxfzl"                     csv:"最新负债率 (%)"`
	// 负债流动比
	FZLDB float64 `json:"fzldb"                     csv:"负债流动比"`
	// 净利润 3 年复合增长率 (%)
	NetprofitGrowthrate3Y float64 `json:"netprofit_growthrate_3_y"  csv:"净利润 3 年复合增长率 (%)"`
	// 营收 3 年复合增长率 (%)
	IncomeGrowthrate3Y float64 `json:"income_growthrate_3_y"     csv:"营收 3 年复合增长率 (%)"`
	// 上市以来年化收益率 (%)
	ListingYieldYear float64 `json:"listing_yield_year"        csv:"上市以来年化收益率 (%)"`
	// 上市以来年化波动率 (%)
	ListingVolatilityYear float64 `json:"listing_volatility_year"   csv:"年化波动率 (%)"`
	// 市盈率
	PE float64 `json:"pe"                        csv:"市盈率"`
	// PEG
	PEG float64 `json:"peg"                       csv:"PEG"`
	// 机构评级
	OrgRating string `json:"org_rating"                csv:"机构评级"`
	// 盈利预测
	ProfitPredict string `json:"profit_predict"            csv:"盈利预测"`
	// 市盈率估值
	ValuationSYL string `json:"valuation_syl"             csv:"市盈率估值"`
	// 市净率估值
	ValuationSJL string `json:"valuation_sjl"             csv:"市净率估值"`
	// 市销率估值
	ValuationSXOL string `json:"valuation_sxol"            csv:" 市销率估值"`
	// 市现率估值
	ValuationSXNL string `json:"valuation_sxnl"            csv:"市现率估值"`
	// 行业均值水平
	HYJZSP string `json:"hyjzsp"                    csv:"行业均值水平"`
	// 整体质地
	ZTZD string `json:"ztzd"                      csv:"整体质地"`
	// 近五年毛利率
	MLL5Y []float64 `json:"mll_5y"                    csv:"近五年毛利率"`
	// 近五年净利率
	JLL5Y []float64 `json:"jll_5y"                    csv:"近五年净利率"`
	// 上市时间
	ListingDate string `json:"listing_date"              csv:"上市时间"`
	// 最新经营活动产生的现金流量净额
	NetcashOperate string `json:"netcash_operate"           csv:"经营现金流净额"`
	// 最新投资活动产生的现金流量净额
	NetcashInvest string `json:"netcash_invest"            csv:"投资现金流净额"`
	// 最新筹资活动产生的现金流量净额
	NetcashFinance string `json:"netcash_finance"           csv:"筹资现金流净额"`
	// 自由现金流
	NetcashFree string `json:"netcash_free"              csv:"自由现金流"`
}

// GetHeaderValueMap 获取以 csv tag 为 key 的 Data map
func (d Data) GetHeaderValueMap() map[string]interface{} {
	return goutils.StructToMap(&d, "csv")
}

// GetHeaders 获取 csv tag 列表
func (d Data) GetHeaders() []string {
	return goutils.StructTagList(&d, "csv")
}

// NewData 创建 Data 对象
func NewData(ctx context.Context, stock model.Stock) Data {
	var rightPrice interface{} = "--"
	var priceSpace interface{} = "--"
	var reportOpinion interface{} = "--"
	if stock.RightPrice > 0 {
		rightPrice = stock.RightPrice
		priceSpace = stock.RightPrice - stock.GetPrice()
	}
	if stock.FinaReportOpinion != "" {
		reportOpinion = stock.FinaReportOpinion
	}

	fina := stock.HistoricalFinaMainData[0]
	return Data{
		Name:            stock.BaseInfo.SecurityNameAbbr,
		Code:            stock.BaseInfo.Secucode,
		Industry:        stock.BaseInfo.Industry,
		Keywords:        stock.CompanyProfile.KeywordsString(),
		CompanyProfile:  stock.CompanyProfile.ProfileString(),
		MainForms:       stock.CompanyProfile.MainFormsString(),
		BYYSRatio:       stock.BYYSRatio,
		ReportDateName:  fina.ReportDateName,
		ReportOpinion:   reportOpinion,
		JZPG:            stock.JZPG.String(),
		LatestROE:       fina.Roejq,
		ROETBZZ:         fina.Roejqtz,
		ROE5Y:           stock.HistoricalFinaMainData.ValueList(ctx, "ROE", 5),
		LatestEPS:       fina.Epsjb,
		EPSTBZZ:         fina.Epsjbtz,
		EPS5Y:           stock.HistoricalFinaMainData.ValueList(ctx, "EPS", 5),
		TotalIncome:     goutils.YiWanString(fina.Totaloperatereve),
		TotalIncomeTBZZ: fina.Totaloperaterevetz,
		TotalIncome5Y:   stock.HistoricalFinaMainData.ValueList(ctx, "REVENUE", 5),

		NetProfit:              goutils.YiWanString(fina.Parentnetprofit),
		NetProfitTBZZ:          fina.Parentnetprofittz,
		NetProfit5Y:            stock.HistoricalFinaMainData.ValueList(ctx, "NETPROFIT", 5),
		ZXGXL:                  stock.BaseInfo.Zxgxl,
		FZLDB:                  fina.Ld,
		FinaReportDate:         strings.Fields(stock.FinaReportDate)[0],
		FinaAppointPublishDate: strings.Fields(stock.FinaAppointPublishDate)[0],
		FinaActualPublishDate:  strings.Fields(stock.FinaActualPublishDate)[0],
		TotalMarketCap:         goutils.YiWanString(stock.BaseInfo.TotalMarketCap),
		Price:                  stock.GetPrice(),
		RightPrice:             rightPrice,
		PriceSpace:             priceSpace,
		HV:                     stock.HistoricalVolatility,
		ListingVolatilityYear:  stock.BaseInfo.ListingVolatilityYear,
		ZXFZL:                  fina.Zcfzl,
		NetprofitGrowthrate3Y:  stock.BaseInfo.NetprofitGrowthrate3Y,
		IncomeGrowthrate3Y:     stock.BaseInfo.IncomeGrowthrate3Y,
		ListingYieldYear:       stock.BaseInfo.ListingYieldYear,
		PE:                     stock.BaseInfo.PE,
		PEG:                    stock.PEG,
		OrgRating:              stock.OrgRatingList.String(),
		ProfitPredict:          stock.ProfitPredictList.String(),
		ValuationSYL:           stock.ValuationMap["市盈率"],
		ValuationSJL:           stock.ValuationMap["市净率"],
		ValuationSXOL:          stock.ValuationMap["市销率"],
		ValuationSXNL:          stock.ValuationMap["市现率"],
		HYJZSP:                 stock.JZPG.GetValuationScore(),
		ZTZD:                   stock.JZPG.GetValueTotalScore(),
		MLL5Y:                  stock.HistoricalFinaMainData.ValueList(ctx, "MLL", 5),
		JLL5Y:                  stock.HistoricalFinaMainData.ValueList(ctx, "JLL", 5),

		ListingDate:    stock.BaseInfo.ListingDate,
		NetcashOperate: goutils.YiWanString(stock.NetcashOperate),
		NetcashInvest:  goutils.YiWanString(stock.NetcashInvest),
		NetcashFinance: goutils.YiWanString(stock.NetcashFinance),
		NetcashFree:    goutils.YiWanString(stock.NetcashFree),
	}
}

// DataList 要导出的数据列表
type DataList []Data

// SortByROE 股票列表按 ROE 排序
func (d DataList) SortByROE() {
	sort.Slice(d, func(i, j int) bool {
		return d[i].LatestROE > d[j].LatestROE
	})
}

// SortByPrice 股票列表按股价排序
func (d DataList) SortByPrice() {
	sort.Slice(d, func(i, j int) bool {
		return d[i].Price < d[j].Price
	})
}

// SortByZXGXL 股票列表按最新股息率排序
func (d DataList) SortByZXGXL() {
	sort.Slice(d, func(i, j int) bool {
		return d[i].ZXGXL > d[j].ZXGXL
	})
}

// SortByHV 股票列表按历史波动率排序
func (d DataList) SortByHV() {
	sort.Slice(d, func(i, j int) bool {
		return d[i].HV > d[j].HV
	})
}

// GetIndustryList 获取行业分类列表
func (d DataList) GetIndustryList() []string {
	result := []string{}
	for _, stock := range d {
		if !goutils.IsStrInSlice(stock.Industry, result) {
			result = append(result, stock.Industry)
		}
	}
	return result
}

// ChunkedBySize 将 stock 列表按大小切割分组
func (d DataList) ChunkedBySize(chunkSize int) []DataList {
	result := []DataList{}
	dataLen := len(d)
	for i := 0; i < dataLen; i += chunkSize {
		end := i + chunkSize
		if end > dataLen {
			end = dataLen
		}
		result = append(result, d[i:end])
	}
	return result
}

// NewDataList 创建要导出的数据列表
func NewDataList(ctx context.Context, stocks model.StockList) (result DataList) {
	for _, s := range stocks {
		result = append(result, NewData(ctx, s))
	}
	return
}
