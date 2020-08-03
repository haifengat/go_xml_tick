package src

import "encoding/xml"

// 实际不使用
type XTPPackages struct {
	XMLName                   xml.Name                    `xml:"XTPPackages"`
	NtfDepthMarketDataPackage []NtfDepthMarketDataPackage `xml:"NtfDepthMarketDataPackage"`
}

type NtfDepthMarketDataPackage struct {
	XMLName                     xml.Name                    `xml:"NtfDepthMarketDataPackage"`
	MarketDataUpdateTimeField   MarketDataUpdateTimeField   `xml:"MarketDataUpdateTimeField,omitempty"`
	MarketDataBaseField         MarketDataBaseField         `xml:"MarketDataBaseField,omitempty"`
	MarketDataStaticField       MarketDataStaticField       `xml:"MarketDataStaticField,omitempty"`
	MarketDataLastMatchField    MarketDataLastMatchField    `xml:"MarketDataLastMatchField,omitempty"`
	MarketDataBestPriceField    MarketDataBestPriceField    `xml:"MarketDataBestPriceField,omitempty"`
	MarketDataAveragePriceField MarketDataAveragePriceField `xml:"MarketDataAveragePriceField,omitempty"`
}

type MarketDataUpdateTimeField struct {
	XMLName        xml.Name `xml:"MarketDataUpdateTimeField"`
	InstrumentID   string   `xml:"InstrumentID,attr,omitempty" db:"InstrumentID"`
	UpdateTime     string   `xml:"UpdateTime,attr" db:"UpdateTime"`
	UpdateMillisec int32    `xml:"UpdateMillisec,attr" db:"UpdateMillisec"`
	ActionDay      string   `xml:"ActionDay,attr,omitempty" db:"ActionDay"`
}

type MarketDataBaseField struct {
	XMLName    xml.Name `xml:"MarketDataBaseField"`
	TradingDay string   `xml:"TradingDay,attr,omitempty" db:"TradingDay"`
}

type MarketDataStaticField struct {
	XMLName         xml.Name `xml:"MarketDataStaticField"`
	UpperLimitPrice float32  `xml:"UpperLimitPrice,attr" db:"UpperLimitPrice"`
	LowerLimitPrice float32  `xml:"LowerLimitPrice,attr" db:"LowerLimitPrice"`
}

type MarketDataLastMatchField struct {
	XMLName      xml.Name `xml:"MarketDataLastMatchField"`
	LastPrice    float32  `xml:"LastPrice,attr" db:"LastPrice"`
	Volume       int32    `xml:"Volume,attr" db:"Volume"`
	Turnover     float32  `xml:"Turnover,attr" db:"Turnover"`
	OpenInterest float32  `xml:"OpenInterest,attr" db:"OpenInterest"`
}

type MarketDataBestPriceField struct {
	XMLName    xml.Name `xml:"MarketDataBestPriceField"`
	BidPrice1  float32  `xml:"BidPrice1,attr" db:"BidPrice1"`
	BidVolume1 int32    `xml:"BidVolume1,attr" db:"BidVolume1"`
	AskPrice1  float32  `xml:"AskPrice1,attr" db:"AskPrice1"`
	AskVolume1 int32    `xml:"AskVolume1,attr" db:"AskVolume1"`
}

type MarketDataAveragePriceField struct {
	XMLName      xml.Name `xml:"MarketDataAveragePriceField"`
	AveragePrice float32  `xml:"AveragePrice,attr" db:"AveragePrice"`
}
