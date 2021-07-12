// 对给定股票进行检测

package core

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/axiaoxin-com/goutils"
	"github.com/axiaoxin-com/logging"
	"github.com/axiaoxin-com/x-stock/datacenter/eastmoney"
	"github.com/axiaoxin-com/x-stock/models"
)

// CheckerOptions 检测条件选项
type CheckerOptions struct {
	// 最新一期 ROE 不低于该值
	MinROE float64 `json:"min_roe"                 form:"checker_min_roe"`
	// 连续增长年数
	CheckYears int `json:"check_years"             form:"checker_check_years"`
	// ROE 高于该值时不做连续增长检查
	NoCheckYearsROE float64 `json:"no_check_years_roe"      form:"checker_no_check_years_roe"`
	// 最大资产负债率百分比(%)
	MaxDebtAssetRatio float64 `json:"max_debt_asset_ratio"    form:"checker_max_debt_asset_ratio"`
	// 最大历史波动率
	MaxHV float64 `json:"max_hv"                  form:"checker_max_hv"`
	// 最小市值（亿）
	MinTotalMarketCap float64 `json:"min_total_market_cap"    form:"checker_min_total_market_cap"`
	// 银行股最小 ROA
	BankMinROA float64 `json:"bank_min_roa"            form:"checker_bank_min_roa"`
	// 银行股最小资本充足率
	BankMinZBCZL float64 `json:"bank_min_zbczl"          form:"checker_bank_min_zbczl"`
	// 银行股最大不良贷款率
	BankMaxBLDKL float64 `json:"bank_max_bldkl"          form:"checker_bank_max_bldkl"`
	// 银行股最低不良贷款拨备覆盖率
	BankMinBLDKBBFGL float64 `json:"bank_min_bldkbbfgl"      form:"checker_bank_min_bldkbbfgl"`
	// 是否检测毛利率稳定性
	IsCheckMLLStability bool `json:"is_check_mll_stability"  form:"checker_is_check_mll_stability"`
	// 是否检测净利率稳定性
	IsCheckJLLStability bool `json:"is_check_jll_stability"  form:"checker_is_check_jll_stability"`
	// 是否使用估算合理价进行检测，高于估算价将被过滤
	IsCheckPriceByCalc bool `json:"is_check_price_by_calc"  form:"checker_is_check_price_by_calc"`
	// 最大 PEG
	MaxPEG float64 `json:"max_peg"                 form:"checker_max_peg"`
	// 最小本业营收比
	MinBYYSRatio float64 `json:"min_byys_ratio"          form:"checker_min_byys_ratio"`
	// 最大本业营收比
	MaxBYYSRatio float64 `json:"max_byys_ratio"          form:"checker_max_byys_ratio"`
	// 最小负债流动比
	MinFZLDB float64 `json:"min_fzldb"               form:"checker_min_fzldb"`
	// 是否检测现金流量
	IsCheckCashflow bool `json:"is_check_cashflow"       form:"checker_is_check_cashflow"`
	// 是否检测毛利率逐年递增
	IsCheckMLLGrow bool `json:"is_check_mll_grow"       form:"checker_is_check_mll_grow"`
	// 是否检测净利率逐年递增
	IsCheckJLLGrow bool `json:"is_check_jll_grow"       form:"checker_is_check_jll_grow"`
	// 是否检测EPS逐年递增
	IsCheckEPSGrow bool `json:"is_check_eps_grow"       form:"checker_is_check_eps_grow"`
	// 是否检测营收逐年递增
	IsCheckRevGrow bool `json:"is_check_rev_grow"       form:"checker_is_check_rev_grow"`
	// 是否检测净利润逐年递增
	IsCheckNetprofitGrow bool `json:"is_check_netprofit_grow" form:"checker_is_check_netprofit_grow"`
	// 最低股息率
	MinGxl float64 `json:"min_gxl"                 form:"checker_min_gxl"`
}

// DefaultCheckerOptions 默认检测值
var DefaultCheckerOptions = CheckerOptions{
	MinROE:               8.0,
	CheckYears:           5,
	NoCheckYearsROE:      20.0,
	MaxDebtAssetRatio:    60.0,
	MaxHV:                1.0,
	MinTotalMarketCap:    100.0,
	BankMinROA:           0.5,
	BankMinZBCZL:         8.0,
	BankMaxBLDKL:         3.0,
	BankMinBLDKBBFGL:     100.0,
	IsCheckJLLStability:  false,
	IsCheckMLLStability:  false,
	IsCheckPriceByCalc:   true,
	MaxPEG:               1.5,
	MinBYYSRatio:         0.9,
	MaxBYYSRatio:         1.1,
	MinFZLDB:             1,
	IsCheckCashflow:      false,
	IsCheckMLLGrow:       false,
	IsCheckJLLGrow:       false,
	IsCheckEPSGrow:       true,
	IsCheckRevGrow:       true,
	IsCheckNetprofitGrow: true,
	MinGxl:               0.0,
}

// Checker 检测器实例
type Checker struct {
	Options CheckerOptions
}

// NewChecker 创建检查器实例
func NewChecker(ctx context.Context, opts CheckerOptions) *Checker {
	return &Checker{
		Options: opts,
	}
}

// CheckResult 检测结果
// key 为检测项，value为描述map {"ROE": {"desc": "高于8.0", "ok":"true"}}
type CheckResult map[string]map[string]string

// CheckFundamentals 检测股票基本面
// [[检测失败项, 原因], ...]
func (c Checker) CheckFundamentals(ctx context.Context, stock models.Stock) (result CheckResult, ok bool) {
	ok = true
	result = make(CheckResult)
	// ROE 高于 n%
	checkItemName := "净资产收益率 (ROE)"
	itemOK := true
	desc := fmt.Sprintf("最新一期 ROE:%f", stock.BaseInfo.RoeWeight)
	if stock.BaseInfo.RoeWeight < c.Options.MinROE {
		desc = fmt.Sprintf("最新一期 ROE:%f 低于:%f", stock.BaseInfo.RoeWeight, c.Options.MinROE)
		ok = false
		itemOK = false
	}
	result[checkItemName] = map[string]string{
		"desc": desc,
		"ok":   fmt.Sprint(itemOK),
	}

	// ROE 均值小于 NoCheckYearsROE 时，至少 n 年内逐年递增
	checkItemName = fmt.Sprintf("ROE 逐年递增（均值>=%f除外）", c.Options.NoCheckYearsROE)
	itemOK = true
	roeList := stock.HistoricalFinaMainData.ValueList(
		ctx,
		eastmoney.ValueListTypeROE,
		c.Options.CheckYears,
		eastmoney.FinaReportTypeYear,
	)
	desc = fmt.Sprintf("%d 年内 ROE:%+v", c.Options.CheckYears, roeList)
	roeavg, err := goutils.AvgFloat64(roeList)
	if err != nil {
		logging.Warn(ctx, "roe avg error:"+err.Error())
	}
	if roeavg < c.Options.NoCheckYearsROE &&
		!stock.HistoricalFinaMainData.IsIncreasingByYears(
			ctx,
			eastmoney.ValueListTypeROE,
			c.Options.CheckYears,
			eastmoney.FinaReportTypeYear,
		) {
		desc = fmt.Sprintf("%d 年内未逐年递增:%+v", c.Options.CheckYears, roeList)
		ok = false
		itemOK = false
	}
	result[checkItemName] = map[string]string{
		"desc": desc,
		"ok":   fmt.Sprint(itemOK),
	}

	// EPS 至少 n 年内逐年递增且 > 0
	checkItemName = "EPS 逐年递增且 > 0"
	itemOK = true
	epsList := stock.HistoricalFinaMainData.ValueList(
		ctx,
		eastmoney.ValueListTypeEPS,
		c.Options.CheckYears,
		eastmoney.FinaReportTypeYear,
	)
	desc = fmt.Sprintf("%d 年内 EPS:%+v", c.Options.CheckYears, epsList)
	if c.Options.IsCheckEPSGrow && len(epsList) > 0 {
		if epsList[len(epsList)-1] <= 0 ||
			!stock.HistoricalFinaMainData.IsIncreasingByYears(
				ctx,
				eastmoney.ValueListTypeEPS,
				c.Options.CheckYears,
				eastmoney.FinaReportTypeYear,
			) {
			ok = false
			itemOK = false
		}
	}
	result[checkItemName] = map[string]string{
		"desc": desc,
		"ok":   fmt.Sprint(itemOK),
	}

	// 营业总收入至少 n 年内逐年递增且 > 0
	revList := stock.HistoricalFinaMainData.ValueList(
		ctx,
		eastmoney.ValueListTypeRevenue,
		c.Options.CheckYears,
		eastmoney.FinaReportTypeYear,
	)
	checkItemName = "营收逐年递增且 > 0"
	itemOK = true
	revs := []string{}
	for _, rev := range revList {
		revs = append(revs, goutils.YiWanString(rev))
	}
	desc = fmt.Sprintf("%d 年内营收: %s", c.Options.CheckYears, strings.Join(revs, ", "))
	if c.Options.IsCheckRevGrow && len(revList) > 0 {
		if revList[len(revList)-1] <= 0 ||
			!stock.HistoricalFinaMainData.IsIncreasingByYears(
				ctx,
				eastmoney.ValueListTypeRevenue,
				c.Options.CheckYears,
				eastmoney.FinaReportTypeYear,
			) {
			ok = false
			itemOK = false
		}
	}
	result[checkItemName] = map[string]string{
		"desc": desc,
		"ok":   fmt.Sprint(itemOK),
	}

	// 净利润至少 n 年内逐年递增
	netprofitList := stock.HistoricalFinaMainData.ValueList(
		ctx,
		eastmoney.ValueListTypeNetProfit,
		c.Options.CheckYears,
		eastmoney.FinaReportTypeYear,
	)
	checkItemName = "净利润逐年递增且 > 0"
	itemOK = true
	nps := []string{}
	for _, np := range netprofitList {
		nps = append(nps, goutils.YiWanString(np))
	}
	desc = fmt.Sprintf("%d 年内净利润: %s", c.Options.CheckYears, strings.Join(nps, ", "))
	if c.Options.IsCheckNetprofitGrow && len(netprofitList) > 0 {
		if netprofitList[len(netprofitList)-1] <= 0 ||
			!stock.HistoricalFinaMainData.IsIncreasingByYears(
				ctx,
				eastmoney.ValueListTypeNetProfit,
				c.Options.CheckYears,
				eastmoney.FinaReportTypeYear,
			) {
			ok = false
			itemOK = false
		}
	}
	result[checkItemName] = map[string]string{
		"desc": desc,
		"ok":   fmt.Sprint(itemOK),
	}

	// 整体质地
	checkItemName = "整体质地"
	itemOK = true
	desc = stock.JZPG.GetValueTotalScore()
	if !goutils.IsStrInSlice(stock.JZPG.GetValueTotalScore(), []string{"优秀", "良好"}) {
		ok = false
		itemOK = false
	}
	result[checkItemName] = map[string]string{
		"desc": desc,
		"ok":   fmt.Sprint(itemOK),
	}

	// 行业均值水平估值
	checkItemName = "行业均值水平估值"
	itemOK = true
	desc = stock.JZPG.GetValuationScore()
	if stock.JZPG.GetValuationScore() == "高于行业均值水平" {
		ok = false
		itemOK = false
	}
	result[checkItemName] = map[string]string{
		"desc": desc,
		"ok":   fmt.Sprint(itemOK),
	}

	// 市盈率、市净率、市现率、市销率全部估值较高
	checkItemName = "四率估值"
	itemOK = true
	allHighValuation := true
	valuationDesc := []string{}
	for k, v := range stock.ValuationMap {
		valuationDesc = append(valuationDesc, k+v)
	}
	for _, v := range stock.ValuationMap {
		if v != "估值较高" {
			allHighValuation = false
			break
		}
	}
	if allHighValuation {
		ok = false
		itemOK = false
	}
	result[checkItemName] = map[string]string{
		"desc": strings.Join(valuationDesc, "\n"),
		"ok":   fmt.Sprint(itemOK),
	}

	// 股价低于合理价格
	checkItemName = "合理股价"
	itemOK = true
	desc = fmt.Sprintf("最新股价:%f 合理价:%f", stock.GetPrice(), stock.RightPrice)
	if c.Options.IsCheckPriceByCalc {
		if stock.RightPrice != -1 && stock.GetPrice() > stock.RightPrice {
			ok = false
			itemOK = false
		}
	}
	result[checkItemName] = map[string]string{
		"desc": desc,
		"ok":   fmt.Sprint(itemOK),
	}

	// 负债率低于 MaxDebtRatio （可选条件），金融股不检测该项
	checkItemName = "负债率"
	itemOK = true
	if len(stock.HistoricalFinaMainData) > 0 {
		fzl := stock.HistoricalFinaMainData[0].Zcfzl
		desc = fmt.Sprintf("负债率:%f", fzl)
		if !goutils.IsStrInSlice(stock.GetOrgType(), []string{"银行", "保险"}) {
			if c.Options.MaxDebtAssetRatio != 0 && len(stock.HistoricalFinaMainData) > 0 {
				if fzl > c.Options.MaxDebtAssetRatio {
					desc = fmt.Sprintf("负债率:%f 高于:%f", fzl, c.Options.MaxDebtAssetRatio)
					ok = false
					itemOK = false
				}
			}
		}
		result[checkItemName] = map[string]string{
			"desc": desc,
			"ok":   fmt.Sprint(itemOK),
		}
	}

	// 历史波动率 （可选条件）
	checkItemName = "历史波动率"
	itemOK = true
	desc = fmt.Sprintf("历史波动率:%f", stock.HistoricalVolatility)
	if c.Options.MaxHV != 0 {
		if stock.HistoricalVolatility > c.Options.MaxHV {
			desc = fmt.Sprintf("历史波动率:%f 高于:%f", stock.HistoricalVolatility, c.Options.MaxHV)
			ok = false
			itemOK = false
		}
	}
	result[checkItemName] = map[string]string{
		"desc": desc,
		"ok":   fmt.Sprint(itemOK),
	}

	// 市值
	checkItemName = "市值"
	itemOK = true
	sz := goutils.YiWanString(stock.BaseInfo.TotalMarketCap)
	desc = fmt.Sprintf("市值:%s", sz)
	if stock.BaseInfo.TotalMarketCap < c.Options.MinTotalMarketCap*100000000 {
		desc = fmt.Sprintf("市值:%s 低于:%f亿", sz, c.Options.MinTotalMarketCap)
		ok = false
		itemOK = false
	}
	result[checkItemName] = map[string]string{
		"desc": desc,
		"ok":   fmt.Sprint(itemOK),
	}

	// 银行股特殊检测
	if stock.GetOrgType() == "银行" && len(stock.HistoricalFinaMainData) > 0 {
		fmd := stock.HistoricalFinaMainData[0]
		checkItemName = "总资产收益率 (ROA)"
		itemOK = true
		desc = fmt.Sprintf("最新 ROA:%f", stock.BaseInfo.ROA)
		if stock.BaseInfo.ROA < c.Options.BankMinROA {
			desc = fmt.Sprintf("ROA:%f 低于:%f", stock.BaseInfo.ROA, c.Options.BankMinROA)
			ok = false
			itemOK = false
		}
		result[checkItemName] = map[string]string{
			"desc": desc,
			"ok":   fmt.Sprint(itemOK),
		}

		checkItemName = "资本充足率"
		itemOK = true
		desc = fmt.Sprintf("资本充足率:%f", fmd.Newcapitalader)
		if fmd.Newcapitalader < c.Options.BankMinZBCZL {
			desc = fmt.Sprintf("资本充足率:%f 低于:%f", fmd.Newcapitalader, c.Options.BankMinZBCZL)
			ok = false
			itemOK = false
		}
		result[checkItemName] = map[string]string{
			"desc": desc,
			"ok":   fmt.Sprint(itemOK),
		}

		checkItemName = "不良贷款率"
		itemOK = true
		desc = fmt.Sprintf("不良贷款率:%f", fmd.Newcapitalader)
		if c.Options.BankMaxBLDKL != 0 {
			if fmd.NonPerLoan > c.Options.BankMaxBLDKL {
				desc = fmt.Sprintf("不良贷款率:%f 高于:%f", fmd.Newcapitalader, c.Options.BankMinZBCZL)
				ok = false
				itemOK = false
			}
		}
		result[checkItemName] = map[string]string{
			"desc": desc,
			"ok":   fmt.Sprint(itemOK),
		}

		checkItemName = "不良贷款拨备覆盖率"
		itemOK = true
		desc = fmt.Sprintf("不良贷款拨备覆盖率:%f", fmd.Newcapitalader)
		if fmd.Bldkbbl < c.Options.BankMinBLDKBBFGL {
			desc = fmt.Sprintf("不良贷款拨备覆盖率:%f 低于:%f", fmd.Newcapitalader, c.Options.BankMinZBCZL)
			ok = false
			itemOK = false
		}
		result[checkItemName] = map[string]string{
			"desc": desc,
			"ok":   fmt.Sprint(itemOK),
		}
	}

	// 毛利率稳定性 （只检测非金融股）
	mllList := stock.HistoricalFinaMainData.ValueList(
		ctx,
		eastmoney.ValueListTypeMLL,
		c.Options.CheckYears,
		eastmoney.FinaReportTypeYear,
	)
	checkItemName = "毛利率稳定性"
	itemOK = true
	desc = fmt.Sprintf("%d 年内毛利率:%v", c.Options.CheckYears, mllList)
	if c.Options.IsCheckMLLStability && !goutils.IsStrInSlice(stock.GetOrgType(), []string{"银行", "保险"}) {
		if !stock.HistoricalFinaMainData.IsStability(
			ctx,
			eastmoney.ValueListTypeMLL,
			c.Options.CheckYears,
			eastmoney.FinaReportTypeYear,
		) {
			desc = fmt.Sprintf("%d 年内稳定性较差:%v", c.Options.CheckYears, mllList)
			ok = false
			itemOK = false
		}
	}
	result[checkItemName] = map[string]string{
		"desc": desc,
		"ok":   fmt.Sprint(itemOK),
	}

	// 毛利率逐年递增
	checkItemName = "毛利率逐年递增且 > 0"
	itemOK = true
	desc = fmt.Sprintf("%d 年内毛利率:%v", c.Options.CheckYears, mllList)
	if c.Options.IsCheckMLLGrow && len(mllList) > 0 {
		if revList[len(mllList)-1] <= 0 ||
			!stock.HistoricalFinaMainData.IsIncreasingByYears(
				ctx,
				eastmoney.ValueListTypeMLL,
				c.Options.CheckYears,
				eastmoney.FinaReportTypeYear,
			) {
			ok = false
			itemOK = false
		}
	}
	result[checkItemName] = map[string]string{
		"desc": desc,
		"ok":   fmt.Sprint(itemOK),
	}

	// 净利率稳定性
	jllList := stock.HistoricalFinaMainData.ValueList(
		ctx,
		eastmoney.ValueListTypeJLL,
		c.Options.CheckYears,
		eastmoney.FinaReportTypeYear,
	)
	checkItemName = "净利率稳定性"
	itemOK = true
	desc = fmt.Sprintf("%d 年内净利率:%v", c.Options.CheckYears, jllList)
	if c.Options.IsCheckJLLStability {
		if !stock.HistoricalFinaMainData.IsStability(
			ctx,
			eastmoney.ValueListTypeJLL,
			c.Options.CheckYears,
			eastmoney.FinaReportTypeYear,
		) {
			desc = fmt.Sprintf("%d 年内稳定性较差:%v", c.Options.CheckYears, jllList)
			ok = false
			itemOK = false
		}
	}
	result[checkItemName] = map[string]string{
		"desc": desc,
		"ok":   fmt.Sprint(itemOK),
	}

	// 净利率逐年递增
	checkItemName = "净利率逐年递增且 > 0"
	itemOK = true
	desc = fmt.Sprintf("%d 年内净利率:%v", c.Options.CheckYears, jllList)
	if c.Options.IsCheckJLLGrow && len(jllList) > 0 {
		if revList[len(jllList)-1] <= 0 ||
			!stock.HistoricalFinaMainData.IsIncreasingByYears(
				ctx,
				eastmoney.ValueListTypeJLL,
				c.Options.CheckYears,
				eastmoney.FinaReportTypeYear,
			) {
			ok = false
			itemOK = false
		}
	}
	result[checkItemName] = map[string]string{
		"desc": desc,
		"ok":   fmt.Sprint(itemOK),
	}

	// PEG
	checkItemName = "PEG"
	itemOK = true
	desc = fmt.Sprintf("PEG:%v", stock.PEG)
	if c.Options.MaxPEG != 0 {
		if stock.PEG > c.Options.MaxPEG {
			desc = fmt.Sprintf("PEG:%v 高于:%v", stock.PEG, c.Options.MaxPEG)
			ok = false
			itemOK = false
		}
	}
	result[checkItemName] = map[string]string{
		"desc": desc,
		"ok":   fmt.Sprint(itemOK),
	}

	// 本业营收比
	checkItemName = "本业营收比"
	itemOK = true
	desc = fmt.Sprintf("当前本业营收比:%v", stock.BYYSRatio)
	if c.Options.MinBYYSRatio != 0 && c.Options.MaxBYYSRatio != 0 {
		if stock.BYYSRatio > c.Options.MaxBYYSRatio || stock.BYYSRatio < c.Options.MinBYYSRatio {
			desc = fmt.Sprintf("当前本业营收比:%v 超出范围:%v-%v", stock.BYYSRatio, c.Options.MinBYYSRatio, c.Options.MaxBYYSRatio)
			ok = false
			itemOK = false
		}
	}
	result[checkItemName] = map[string]string{
		"desc": desc,
		"ok":   fmt.Sprint(itemOK),
	}

	// 审计意见
	checkItemName = "财报审计意见"
	itemOK = true
	desc = stock.FinaReportOpinion
	if stock.FinaReportOpinion != "" {
		if stock.FinaReportOpinion != "标准无保留意见" {
			ok = false
			itemOK = false
		}
	}
	result[checkItemName] = map[string]string{
		"desc": desc,
		"ok":   fmt.Sprint(itemOK),
	}

	// 配发股利股息
	checkItemName = "配发股利股息"
	itemOK = true
	desc = fmt.Sprintf("最新股息率: %f", stock.BaseInfo.Zxgxl)
	if stock.BaseInfo.Zxgxl < c.Options.MinGxl {
		desc = fmt.Sprintf("最新股息率: %f < %f", stock.BaseInfo.Zxgxl, c.Options.MinGxl)
		ok = false
		itemOK = false
	}
	result[checkItemName] = map[string]string{
		"desc": desc,
		"ok":   fmt.Sprint(itemOK),
	}

	// 负债流动比检测
	checkItemName = "负债流动比"
	itemOK = true
	if len(stock.HistoricalFinaMainData) > 0 {
		fzldb := stock.HistoricalFinaMainData[0].Ld
		desc = fmt.Sprintf("最新负债流动比: %f", fzldb)
		if fzldb < c.Options.MinFZLDB {
			ok = false
			itemOK = false
		}
		result[checkItemName] = map[string]string{
			"desc": desc,
			"ok":   fmt.Sprint(itemOK),
		}
	}

	// 现金流检测
	checkItemName = "现金流量"
	itemOK = true
	if len(stock.HistoricalCashflowList) > 0 {
		desc = fmt.Sprintf(
			`经营活动产生的现金流量净额(>0): %s; 投资活动产生的现金流量净额(<0)：%s; 筹资活动产生的现金流量净额: %s; 自由现金流量(>0)：%s`,
			goutils.YiWanString(stock.NetcashOperate),
			goutils.YiWanString(stock.NetcashInvest),
			goutils.YiWanString(stock.NetcashFinance),
			goutils.YiWanString(stock.NetcashFree),
		)
		if c.Options.IsCheckCashflow {
			if stock.NetcashOperate < 0 {
				ok = false
				itemOK = false
			}
			if stock.NetcashInvest > 0 {
				ok = false
				itemOK = false
			}
			if stock.NetcashFree < 0 {
				ok = false
				itemOK = false
			}
		}
		result[checkItemName] = map[string]string{
			"desc": desc,
			"ok":   fmt.Sprint(itemOK),
		}
	}

	return
}

// FundStocksCheckResult 股票持仓检测结果
type FundStocksCheckResult struct {
	Names                   []string      `json:"names"`
	CheckResults            []CheckResult `json:"check_results"`
	FinaReportNames         []string      `json:"fina_report_names"`
	FinaAppointPublishDates []string      `json:"fina_appoint_publish_dates"`
}

// CheckFundStocks 检测基金持仓股票
// 返回结果 {"stockname": {"ROE": "xx", "EPS": ""}}
func (c Checker) CheckFundStocks(ctx context.Context, fund *models.Fund) (results FundStocksCheckResult, err error) {
	var wg sync.WaitGroup
	wg.Add(len(fund.Stocks))
	codes := []string{}
	for _, s := range fund.Stocks {
		codes = append(codes, s.Code)
	}
	searcher := NewSearcher(ctx)
	stocks, err := searcher.SearchStocks(ctx, codes)
	if err != nil {
		return
	}
	for _, stock := range stocks {
		result, _ := c.CheckFundamentals(ctx, stock)
		name := fmt.Sprintf("%s-%s", stock.BaseInfo.SecurityNameAbbr, stock.BaseInfo.Secucode)
		results.Names = append(results.Names, name)
		results.CheckResults = append(results.CheckResults, result)
		finaReportName := ""
		if len(stock.HistoricalFinaMainData) > 0 {
			finaReportName = stock.HistoricalFinaMainData[0].ReportDateName
		}
		results.FinaReportNames = append(results.FinaReportNames, finaReportName)
		results.FinaAppointPublishDates = append(
			results.FinaAppointPublishDates,
			strings.Split(stock.FinaAppointPublishDate, " ")[0],
		)
	}
	return
}
