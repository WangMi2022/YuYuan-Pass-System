package main

import (
	"context"
	"database/sql"
	"errors"
	"flag"
	"fmt"
	"log"
	"math"
	"math/rand"
	"net"
	"net/url"
	"os"
	"regexp"
	"sort"
	"strings"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

type dbConfig struct {
	Host     string
	Port     string
	Database string
	Username string
	Password string
	Config   string
}

type category struct {
	ID   int64
	Name string
	Code string
}

type assetTemplate struct {
	Name      string
	Brand     string
	Model     string
	Unit      string
	BasePrice float64
	MinQty    int
	MaxQty    int
	WarrantyY int
	Supplier  string
	Location  string
	Custodian string
}

type seededAsset struct {
	ID        int64
	Code      string
	Name      string
	Quantity  int
	Status    string
	Location  string
	Custodian string
	CreatedAt time.Time
}

var defaultCategories = []struct {
	Name        string
	Code        string
	Color       string
	Sort        int
	Description string
}{
	{Name: "座椅/板凳", Code: "FURN-CHAIR", Color: "#0F766E", Sort: 10, Description: "办公椅、会议椅、板凳等座具"},
	{Name: "桌类家具", Code: "FURN-DESK", Color: "#7C3AED", Sort: 20, Description: "办公桌、会议桌、工作台等"},
	{Name: "柜架家具", Code: "FURN-STORAGE", Color: "#9333EA", Sort: 25, Description: "文件柜、密集架、工具柜及储物设施"},
	{Name: "电脑整机", Code: "IT-COMPUTER", Color: "#2563EB", Sort: 30, Description: "台式机、笔记本、工作站等"},
	{Name: "显示设备", Code: "IT-DISPLAY", Color: "#0891B2", Sort: 40, Description: "显示器、电视、投影仪及大屏"},
	{Name: "网络设备", Code: "IT-NETWORK", Color: "#4F46E5", Sort: 50, Description: "交换机、路由器、防火墙及无线设备"},
	{Name: "服务器与存储", Code: "IT-SERVER", Color: "#1D4ED8", Sort: 55, Description: "服务器、存储阵列、备份设备及机柜"},
	{Name: "办公设备", Code: "OFFICE-EQUIP", Color: "#D97706", Sort: 60, Description: "打印机、扫描仪、碎纸机等"},
	{Name: "生产设备", Code: "PROD-EQUIP", Color: "#DC2626", Sort: 70, Description: "生产、检测、维修相关设备"},
	{Name: "安防设备", Code: "SECURITY", Color: "#BE123C", Sort: 75, Description: "门禁、监控、报警及巡检设备"},
	{Name: "机电设施", Code: "FACILITY", Color: "#0F766E", Sort: 80, Description: "空调、供配电、消防及环境保障设施"},
	{Name: "车辆运输", Code: "VEHICLE", Color: "#B45309", Sort: 85, Description: "公务车辆、叉车、搬运车及运输装备"},
	{Name: "软件与许可", Code: "SOFTWARE", Color: "#0369A1", Sort: 90, Description: "商业软件、订阅许可及数字化无形资产"},
	{Name: "其他资产", Code: "OTHER", Color: "#475569", Sort: 99, Description: "暂未归入其他分类的资产"},
}

var templatesByCode = map[string][]assetTemplate{
	"FURN-CHAIR": {
		{Name: "人体工学办公椅", Brand: "永艺", Model: "XY-Ergo Pro", Unit: "把", BasePrice: 860, MinQty: 2, MaxQty: 12, WarrantyY: 3, Supplier: "上海办公家具供应中心", Location: "一号办公楼 3F", Custodian: "行政部-李娜"},
		{Name: "会议椅", Brand: "震旦", Model: "AURORA MC-18", Unit: "把", BasePrice: 420, MinQty: 6, MaxQty: 24, WarrantyY: 2, Supplier: "杭州震旦家具服务商", Location: "会议中心 A 区", Custodian: "行政部-王磊"},
		{Name: "培训折叠椅", Brand: "圣奥", Model: "SUNON Fold-C", Unit: "把", BasePrice: 210, MinQty: 8, MaxQty: 30, WarrantyY: 2, Supplier: "浙江圣奥办公家具", Location: "培训教室 2F", Custodian: "培训中心-陈晨"},
		{Name: "吧台高脚凳", Brand: "宜家", Model: "FRANKLIN", Unit: "把", BasePrice: 299, MinQty: 2, MaxQty: 10, WarrantyY: 1, Supplier: "宜家企业购", Location: "员工休闲区", Custodian: "行政部-赵敏"},
	},
	"FURN-DESK": {
		{Name: "L 型办公桌", Brand: "震旦", Model: "AURORA LD-160", Unit: "张", BasePrice: 1680, MinQty: 1, MaxQty: 6, WarrantyY: 5, Supplier: "杭州震旦家具服务商", Location: "一号办公楼 2F", Custodian: "行政部-李娜"},
		{Name: "升降办公桌", Brand: "乐歌", Model: "E5-HD", Unit: "张", BasePrice: 2390, MinQty: 1, MaxQty: 5, WarrantyY: 5, Supplier: "宁波乐歌人体工学", Location: "研发中心 5F", Custodian: "研发部-周航"},
		{Name: "12 人会议桌", Brand: "圣奥", Model: "SUNON MT-480", Unit: "张", BasePrice: 8600, MinQty: 1, MaxQty: 2, WarrantyY: 5, Supplier: "浙江圣奥办公家具", Location: "会议中心 B 区", Custodian: "行政部-王磊"},
		{Name: "移动工作台", Brand: "海太欧林", Model: "ONLEAD WS-140", Unit: "张", BasePrice: 1250, MinQty: 1, MaxQty: 8, WarrantyY: 3, Supplier: "广州海太欧林", Location: "运维实验室", Custodian: "运维部-孙杰"},
	},
	"FURN-STORAGE": {
		{Name: "五节文件柜", Brand: "震旦", Model: "AURORA FC-05", Unit: "组", BasePrice: 1680, MinQty: 1, MaxQty: 5, WarrantyY: 5, Supplier: "杭州震旦家具服务商", Location: "档案室", Custodian: "档案室-唐宁"},
		{Name: "智能密集架", Brand: "花都", Model: "HD-MJ-12", Unit: "列", BasePrice: 26800, MinQty: 1, MaxQty: 3, WarrantyY: 5, Supplier: "洛阳花都家具集团", Location: "档案库 B 区", Custodian: "档案室-唐宁"},
		{Name: "重型工具柜", Brand: "天钢", Model: "TANKO TG-180", Unit: "台", BasePrice: 5200, MinQty: 1, MaxQty: 4, WarrantyY: 5, Supplier: "工业工位器具供应商", Location: "维修工位", Custodian: "设备部-顾维"},
		{Name: "移动储物柜", Brand: "圣奥", Model: "SUNON MC-12", Unit: "组", BasePrice: 1280, MinQty: 2, MaxQty: 8, WarrantyY: 3, Supplier: "浙江圣奥办公家具", Location: "一号办公楼 2F", Custodian: "行政部-李娜"},
	},
	"IT-COMPUTER": {
		{Name: "开发笔记本", Brand: "联想", Model: "ThinkPad T14p", Unit: "台", BasePrice: 8999, MinQty: 1, MaxQty: 4, WarrantyY: 3, Supplier: "联想企业购", Location: "研发中心 6F", Custodian: "研发部-周航"},
		{Name: "办公台式机", Brand: "戴尔", Model: "OptiPlex 7010", Unit: "台", BasePrice: 5299, MinQty: 1, MaxQty: 8, WarrantyY: 3, Supplier: "戴尔商用授权代理", Location: "一号办公楼 4F", Custodian: "信息部-刘洋"},
		{Name: "图形工作站", Brand: "惠普", Model: "Z2 Tower G9", Unit: "台", BasePrice: 16800, MinQty: 1, MaxQty: 2, WarrantyY: 3, Supplier: "惠普企业解决方案", Location: "设计中心", Custodian: "设计部-林可"},
		{Name: "工控机", Brand: "研华", Model: "IPC-610L", Unit: "台", BasePrice: 7200, MinQty: 1, MaxQty: 3, WarrantyY: 3, Supplier: "研华科技渠道商", Location: "生产车间 A 区", Custodian: "生产部-胡强"},
	},
	"IT-DISPLAY": {
		{Name: "27 寸办公显示器", Brand: "戴尔", Model: "U2723QE", Unit: "台", BasePrice: 3699, MinQty: 1, MaxQty: 8, WarrantyY: 3, Supplier: "戴尔商用授权代理", Location: "研发中心 6F", Custodian: "信息部-刘洋"},
		{Name: "会议交互大屏", Brand: "MAXHUB", Model: "CF86MA", Unit: "台", BasePrice: 29800, MinQty: 1, MaxQty: 1, WarrantyY: 3, Supplier: "视源股份服务商", Location: "会议中心 A 区", Custodian: "行政部-王磊"},
		{Name: "激光投影仪", Brand: "爱普生", Model: "CB-L530U", Unit: "台", BasePrice: 13200, MinQty: 1, MaxQty: 2, WarrantyY: 3, Supplier: "爱普生授权经销商", Location: "培训教室 2F", Custodian: "培训中心-陈晨"},
		{Name: "监控显示屏", Brand: "海康威视", Model: "DS-D50UK55", Unit: "台", BasePrice: 4200, MinQty: 1, MaxQty: 4, WarrantyY: 3, Supplier: "海康威视本地服务商", Location: "安防监控室", Custodian: "安保部-钱峰"},
	},
	"IT-NETWORK": {
		{Name: "核心交换机", Brand: "华三", Model: "S7506X", Unit: "台", BasePrice: 68500, MinQty: 1, MaxQty: 1, WarrantyY: 5, Supplier: "新华三金牌代理", Location: "数据机房 A 区", Custodian: "网络组-马超"},
		{Name: "接入交换机", Brand: "华为", Model: "S5735S-L24T4S", Unit: "台", BasePrice: 4200, MinQty: 1, MaxQty: 6, WarrantyY: 3, Supplier: "华为企业网络代理", Location: "弱电间 3F", Custodian: "网络组-马超"},
		{Name: "下一代防火墙", Brand: "深信服", Model: "AF-1000-B1150", Unit: "台", BasePrice: 36000, MinQty: 1, MaxQty: 1, WarrantyY: 3, Supplier: "深信服安全服务商", Location: "数据机房 A 区", Custodian: "安全组-沈靖"},
		{Name: "无线 AP", Brand: "锐捷", Model: "RG-AP820-L", Unit: "台", BasePrice: 1180, MinQty: 2, MaxQty: 12, WarrantyY: 3, Supplier: "锐捷网络代理", Location: "一号办公楼 1F", Custodian: "网络组-马超"},
	},
	"IT-SERVER": {
		{Name: "虚拟化服务器", Brand: "戴尔", Model: "PowerEdge R760", Unit: "台", BasePrice: 128000, MinQty: 1, MaxQty: 2, WarrantyY: 5, Supplier: "戴尔数据中心解决方案", Location: "数据机房 A 区", Custodian: "运维部-孙杰"},
		{Name: "全闪存储阵列", Brand: "华为", Model: "OceanStor Dorado 3000", Unit: "套", BasePrice: 286000, MinQty: 1, MaxQty: 1, WarrantyY: 5, Supplier: "华为存储金牌代理", Location: "数据机房 A 区", Custodian: "运维部-孙杰"},
		{Name: "备份一体机", Brand: "深信服", Model: "aStor-Backup 2100", Unit: "台", BasePrice: 98000, MinQty: 1, MaxQty: 1, WarrantyY: 3, Supplier: "深信服云计算服务商", Location: "容灾机房", Custodian: "安全组-沈靖"},
		{Name: "标准服务器机柜", Brand: "图腾", Model: "G3.6042", Unit: "台", BasePrice: 6800, MinQty: 1, MaxQty: 4, WarrantyY: 5, Supplier: "机房工程集成商", Location: "数据机房 B 区", Custodian: "运维部-孙杰"},
	},
	"OFFICE-EQUIP": {
		{Name: "黑白激光打印机", Brand: "惠普", Model: "LaserJet Pro M405d", Unit: "台", BasePrice: 2599, MinQty: 1, MaxQty: 5, WarrantyY: 2, Supplier: "办公设备集采平台", Location: "一号办公楼 4F", Custodian: "行政部-赵敏"},
		{Name: "高速扫描仪", Brand: "富士通", Model: "fi-8150", Unit: "台", BasePrice: 6200, MinQty: 1, MaxQty: 2, WarrantyY: 2, Supplier: "办公设备集采平台", Location: "档案室", Custodian: "档案室-唐宁"},
		{Name: "多功能复合机", Brand: "佳能", Model: "imageRUNNER C3226", Unit: "台", BasePrice: 18800, MinQty: 1, MaxQty: 1, WarrantyY: 3, Supplier: "佳能办公渠道", Location: "文印室", Custodian: "行政部-赵敏"},
		{Name: "碎纸机", Brand: "得力", Model: "GA701", Unit: "台", BasePrice: 899, MinQty: 1, MaxQty: 4, WarrantyY: 1, Supplier: "得力企业购", Location: "财务室", Custodian: "财务部-黄洁"},
	},
	"PROD-EQUIP": {
		{Name: "数字示波器", Brand: "Keysight", Model: "DSOX1204G", Unit: "台", BasePrice: 28500, MinQty: 1, MaxQty: 2, WarrantyY: 3, Supplier: "是德科技授权代理", Location: "电子实验室", Custodian: "实验室-韩工"},
		{Name: "工业条码打印机", Brand: "斑马", Model: "ZT411", Unit: "台", BasePrice: 9600, MinQty: 1, MaxQty: 3, WarrantyY: 2, Supplier: "工业设备集成商", Location: "生产车间 B 区", Custodian: "生产部-胡强"},
		{Name: "电动扭矩工具", Brand: "博世", Model: "GDS 18V-1050 H", Unit: "套", BasePrice: 3600, MinQty: 1, MaxQty: 6, WarrantyY: 2, Supplier: "博世工业工具代理", Location: "维修工位", Custodian: "设备部-顾维"},
		{Name: "环境检测仪", Brand: "福禄克", Model: "975V", Unit: "台", BasePrice: 7200, MinQty: 1, MaxQty: 3, WarrantyY: 2, Supplier: "福禄克仪器代理", Location: "质检室", Custodian: "质检部-郑欣"},
	},
	"SECURITY": {
		{Name: "智能门禁终端", Brand: "海康威视", Model: "DS-K1T671", Unit: "台", BasePrice: 2450, MinQty: 1, MaxQty: 6, WarrantyY: 2, Supplier: "海康威视本地服务商", Location: "一号办公楼大厅", Custodian: "安保部-钱峰"},
		{Name: "网络摄像机", Brand: "大华", Model: "DH-IPC-HFW5443M", Unit: "台", BasePrice: 1650, MinQty: 4, MaxQty: 16, WarrantyY: 3, Supplier: "大华安防集成商", Location: "园区外围", Custodian: "安保部-钱峰"},
		{Name: "硬盘录像机", Brand: "海康威视", Model: "DS-8664N-I16", Unit: "台", BasePrice: 9800, MinQty: 1, MaxQty: 2, WarrantyY: 3, Supplier: "海康威视本地服务商", Location: "安防监控室", Custodian: "安保部-钱峰"},
		{Name: "电子巡更终端", Brand: "蓝卡", Model: "BP-2012S", Unit: "套", BasePrice: 3200, MinQty: 1, MaxQty: 4, WarrantyY: 2, Supplier: "园区安防服务商", Location: "安保值班室", Custodian: "安保部-钱峰"},
	},
	"FACILITY": {
		{Name: "精密空调", Brand: "艾默生", Model: "P1025FA", Unit: "台", BasePrice: 86000, MinQty: 1, MaxQty: 3, WarrantyY: 5, Supplier: "数据中心机电服务商", Location: "数据机房 A 区", Custodian: "设备部-顾维"},
		{Name: "模块化 UPS", Brand: "科华", Model: "YTM33100", Unit: "套", BasePrice: 118000, MinQty: 1, MaxQty: 2, WarrantyY: 5, Supplier: "科华数据授权服务商", Location: "配电室", Custodian: "设备部-顾维"},
		{Name: "新风机组", Brand: "美的", Model: "MDV-D140T2", Unit: "台", BasePrice: 23500, MinQty: 1, MaxQty: 4, WarrantyY: 3, Supplier: "园区机电维保单位", Location: "一号办公楼屋面", Custodian: "物业部-许峰"},
		{Name: "消防控制主机", Brand: "海湾", Model: "JB-QG-GST5000", Unit: "台", BasePrice: 36800, MinQty: 1, MaxQty: 2, WarrantyY: 5, Supplier: "消防工程服务商", Location: "消防控制室", Custodian: "物业部-许峰"},
	},
	"VEHICLE": {
		{Name: "新能源公务车", Brand: "比亚迪", Model: "汉 EV", Unit: "辆", BasePrice: 229800, MinQty: 1, MaxQty: 2, WarrantyY: 6, Supplier: "比亚迪企业客户中心", Location: "园区地下车库", Custodian: "行政部-王磊"},
		{Name: "电动叉车", Brand: "杭叉", Model: "XC2.5", Unit: "辆", BasePrice: 126000, MinQty: 1, MaxQty: 2, WarrantyY: 3, Supplier: "杭叉集团区域经销商", Location: "物流仓库", Custodian: "物流部-高鹏"},
		{Name: "电动搬运车", Brand: "林德", Model: "T20SP", Unit: "辆", BasePrice: 56800, MinQty: 1, MaxQty: 3, WarrantyY: 3, Supplier: "林德物料搬运服务商", Location: "生产车间 B 区", Custodian: "物流部-高鹏"},
		{Name: "园区巡逻车", Brand: "玛西尔", Model: "DN-8", Unit: "辆", BasePrice: 48600, MinQty: 1, MaxQty: 2, WarrantyY: 3, Supplier: "园区车辆服务商", Location: "安保停车区", Custodian: "安保部-钱峰"},
	},
	"SOFTWARE": {
		{Name: "三维设计软件许可", Brand: "达索", Model: "SOLIDWORKS Professional", Unit: "套", BasePrice: 68000, MinQty: 1, MaxQty: 4, WarrantyY: 1, Supplier: "达索系统授权代理", Location: "软件资产池", Custodian: "设计部-林可"},
		{Name: "数据库企业版许可", Brand: "Microsoft", Model: "SQL Server Enterprise", Unit: "套", BasePrice: 108000, MinQty: 1, MaxQty: 2, WarrantyY: 1, Supplier: "微软云解决方案提供商", Location: "软件资产池", Custodian: "运维部-孙杰"},
		{Name: "终端安全许可", Brand: "奇安信", Model: "天擎终端安全", Unit: "节点", BasePrice: 680, MinQty: 20, MaxQty: 80, WarrantyY: 1, Supplier: "奇安信安全服务商", Location: "软件资产池", Custodian: "安全组-沈靖"},
		{Name: "项目管理平台许可", Brand: "Atlassian", Model: "Jira Software Data Center", Unit: "套", BasePrice: 76000, MinQty: 1, MaxQty: 1, WarrantyY: 1, Supplier: "企业软件授权服务商", Location: "软件资产池", Custodian: "信息部-刘洋"},
	},
	"OTHER": {
		{Name: "智能门禁终端", Brand: "海康威视", Model: "DS-K1T671", Unit: "台", BasePrice: 2450, MinQty: 1, MaxQty: 5, WarrantyY: 2, Supplier: "海康威视本地服务商", Location: "一号办公楼大厅", Custodian: "安保部-钱峰"},
		{Name: "UPS 电源", Brand: "山特", Model: "C3KS", Unit: "台", BasePrice: 5200, MinQty: 1, MaxQty: 4, WarrantyY: 3, Supplier: "山特电源渠道商", Location: "数据机房 A 区", Custodian: "运维部-孙杰"},
		{Name: "移动白板", Brand: "得力", Model: "7884", Unit: "块", BasePrice: 580, MinQty: 1, MaxQty: 6, WarrantyY: 1, Supplier: "得力企业购", Location: "培训教室 2F", Custodian: "培训中心-陈晨"},
		{Name: "空气净化器", Brand: "小米", Model: "Air Purifier 4 Pro", Unit: "台", BasePrice: 1499, MinQty: 1, MaxQty: 8, WarrantyY: 1, Supplier: "小米企业购", Location: "员工休闲区", Custodian: "行政部-赵敏"},
	},
}

var genericTemplates = []assetTemplate{
	{Name: "通用资产", Brand: "国产", Model: "GEN-100", Unit: "件", BasePrice: 1000, MinQty: 1, MaxQty: 5, WarrantyY: 2, Supplier: "综合供应商", Location: "综合仓库", Custodian: "资产管理员"},
	{Name: "备用设备", Brand: "国产", Model: "SPARE-200", Unit: "件", BasePrice: 1500, MinQty: 1, MaxQty: 4, WarrantyY: 2, Supplier: "综合供应商", Location: "备件库", Custodian: "资产管理员"},
}

func main() {
	var (
		configPath = flag.String("config", "../deploy/docker-dev/config.yaml", "gin-vue-admin config.yaml path")
		count      = flag.Int("count", 100, "number of fake assets to seed")
		prefix     = flag.String("prefix", "DEMO-ASSET", "asset code prefix for generated data")
		reset      = flag.Bool("reset", true, "delete existing generated assets with the same prefix before seeding")
		seed       = flag.Int64("seed", 20260715, "deterministic random seed")
	)
	flag.Parse()

	if *count <= 0 {
		log.Fatalf("count must be greater than 0")
	}
	*prefix = strings.ToUpper(strings.TrimSpace(*prefix))
	if *prefix == "" {
		log.Fatalf("prefix cannot be empty")
	}

	cfg, err := readDBConfig(*configPath)
	if err != nil {
		log.Fatalf("read config: %v", err)
	}
	db, err := sql.Open("pgx", cfg.DSN())
	if err != nil {
		log.Fatalf("open database: %v", err)
	}
	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		log.Fatalf("connect database %s:%s/%s failed: %v", cfg.Host, cfg.Port, cfg.Database, err)
	}

	if err := ensureSchema(ctx, db); err != nil {
		log.Fatalf("ensure schema: %v", err)
	}
	if err := ensureDefaultCategories(ctx, db); err != nil {
		log.Fatalf("ensure categories: %v", err)
	}

	cats, err := loadCategories(ctx, db)
	if err != nil {
		log.Fatalf("load categories: %v", err)
	}
	if len(cats) == 0 {
		log.Fatalf("no enabled asset categories found")
	}

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		log.Fatalf("begin transaction: %v", err)
	}
	defer tx.Rollback()

	if *reset {
		if _, err := tx.ExecContext(ctx, `
DELETE FROM asset_operation_records
WHERE asset_id IN (SELECT id FROM assets WHERE asset_code LIKE $1)`, *prefix+"-%"); err != nil {
			log.Fatalf("reset generated operation records: %v", err)
		}
		if _, err := tx.ExecContext(ctx, `
DELETE FROM asset_operation_items
WHERE asset_id IN (SELECT id FROM assets WHERE asset_code LIKE $1)`, *prefix+"-%"); err != nil {
			log.Fatalf("reset generated operation items: %v", err)
		}
		if _, err := tx.ExecContext(ctx, `
DELETE FROM asset_operation_orders
WHERE order_no LIKE $1`, "SEED-"+sanitizeCode(*prefix)+"-%"); err != nil {
			log.Fatalf("reset generated operation orders: %v", err)
		}
		if _, err := tx.ExecContext(ctx, `DELETE FROM assets WHERE asset_code LIKE $1`, *prefix+"-%"); err != nil {
			log.Fatalf("reset generated assets: %v", err)
		}
	}

	statsByCategory := map[string]int{}
	statsByStatus := map[string]int{}
	rng := rand.New(rand.NewSource(*seed))
	perCategoryIndex := map[string]int{}
	now := time.Now().Truncate(time.Second)
	seededAssets := make([]seededAsset, 0, *count)

	insertSQL := `
INSERT INTO assets (
  created_at, updated_at, asset_code, name, category_id, brand, model, serial_number,
  quantity, unit, unit_price, original_value, current_value, status, location, custodian,
  supplier, purchase_date, warranty_end_date, photos, remarks
) VALUES (
  $1, $2, $3, $4, $5, $6, $7, $8,
  $9, $10, $11, $12, $13, $14, $15, $16,
  $17, $18, $19, $20::jsonb, $21
)
ON CONFLICT (asset_code) DO UPDATE SET
  updated_at = EXCLUDED.updated_at,
  name = EXCLUDED.name,
  category_id = EXCLUDED.category_id,
  brand = EXCLUDED.brand,
  model = EXCLUDED.model,
  serial_number = EXCLUDED.serial_number,
  quantity = EXCLUDED.quantity,
  unit = EXCLUDED.unit,
  unit_price = EXCLUDED.unit_price,
  original_value = EXCLUDED.original_value,
  current_value = EXCLUDED.current_value,
  status = EXCLUDED.status,
  location = EXCLUDED.location,
  custodian = EXCLUDED.custodian,
  supplier = EXCLUDED.supplier,
  purchase_date = EXCLUDED.purchase_date,
  warranty_end_date = EXCLUDED.warranty_end_date,
  photos = EXCLUDED.photos,
  remarks = EXCLUDED.remarks,
  deleted_at = NULL
RETURNING id`

	for i := 0; i < *count; i++ {
		cat := cats[i%len(cats)]
		tpl := pickTemplate(cat.Code, perCategoryIndex[cat.Code])
		perCategoryIndex[cat.Code]++

		quantity := tpl.MinQty
		if tpl.MaxQty > tpl.MinQty {
			quantity += rng.Intn(tpl.MaxQty - tpl.MinQty + 1)
		}
		unitPrice := money(tpl.BasePrice * (0.88 + rng.Float64()*0.26))
		originalValue := money(float64(quantity) * unitPrice)
		status := statusForIndex(i)
		purchaseDate := now.AddDate(-rng.Intn(5), -rng.Intn(12), -rng.Intn(26))
		if status == "pending_inbound" {
			purchaseDate = now.AddDate(0, 0, -(3 + rng.Intn(24)))
		}
		warrantyEnd := purchaseDate.AddDate(tpl.WarrantyY, 0, 0)
		currentValue := depreciatedValue(originalValue, purchaseDate, status, rng)
		createdAt := now.Add(-time.Duration((*count-i)*9) * time.Hour)
		updatedAt := createdAt.Add(time.Duration(rng.Intn(240)) * time.Hour)
		if updatedAt.After(now) {
			updatedAt = now
		}

		assetCode := fmt.Sprintf("%s-%03d", *prefix, i+1)
		assetName := fmt.Sprintf("%s-%02d", tpl.Name, perCategoryIndex[cat.Code])
		serial := fmt.Sprintf("SN-%s-%04d-%03d", sanitizeCode(cat.Code), purchaseDate.Year(), i+1)
		location, custodian := lifecycleAssignment(tpl, status, i)
		remarks := lifecycleRemark(cat.Name, status, location)

		var assetID int64
		if err := tx.QueryRowContext(ctx, insertSQL,
			createdAt, updatedAt, assetCode, assetName, cat.ID, tpl.Brand, tpl.Model, serial,
			quantity, tpl.Unit, unitPrice, originalValue, currentValue, status, location, custodian,
			tpl.Supplier, purchaseDate.Format("2006-01-02"), warrantyEnd.Format("2006-01-02"), "[]", remarks,
		).Scan(&assetID); err != nil {
			log.Fatalf("insert %s failed: %v", assetCode, err)
		}
		seededAssets = append(seededAssets, seededAsset{
			ID: assetID, Code: assetCode, Name: assetName, Quantity: quantity,
			Status: status, Location: location, Custodian: custodian, CreatedAt: createdAt,
		})

		statsByCategory[cat.Name]++
		statsByStatus[status]++
	}

	operationStats, err := seedOperationData(ctx, tx, seededAssets, *prefix, now)
	if err != nil {
		log.Fatalf("seed asset operation data: %v", err)
	}
	draftStats, err := seedOperationDrafts(ctx, tx, seededAssets, *prefix, now, 2)
	if err != nil {
		log.Fatalf("seed editable operation drafts: %v", err)
	}

	if err := tx.Commit(); err != nil {
		log.Fatalf("commit: %v", err)
	}

	fmt.Printf("已生成资产测试数据：%d 条，前缀：%s，数据库：%s:%s/%s\n", *count, *prefix, cfg.Host, cfg.Port, cfg.Database)
	fmt.Println("分类覆盖：")
	for _, name := range sortedKeys(statsByCategory) {
		fmt.Printf("  - %s：%d 条\n", name, statsByCategory[name])
	}
	fmt.Println("状态分布：")
	for _, status := range []string{"pending_inbound", "idle", "in_use", "maintenance", "retired"} {
		fmt.Printf("  - %s：%d 条\n", statusLabel(status), statsByStatus[status])
	}
	fmt.Println("生命周期单据（已完成）：")
	for _, operationType := range []string{"inbound", "issue", "transfer", "return", "maintenance", "scrap"} {
		fmt.Printf("  - %s：%d 张\n", operationLabel(operationType), operationStats[operationType])
	}
	fmt.Println("可编辑草稿：")
	for _, operationType := range []string{"inbound", "issue", "transfer", "return", "maintenance", "scrap"} {
		fmt.Printf("  - %s：%d 张\n", operationLabel(operationType), draftStats[operationType])
	}
}

func readDBConfig(path string) (dbConfig, error) {
	text, err := os.ReadFile(path)
	if err != nil {
		return dbConfig{}, err
	}
	values := map[string]string{}
	inPgSQL := false
	for _, raw := range strings.Split(string(text), "\n") {
		line := strings.TrimRight(raw, " \t")
		if strings.TrimSpace(line) == "pgsql:" {
			inPgSQL = true
			continue
		}
		if inPgSQL && line != "" && !strings.HasPrefix(line, " ") && !strings.HasPrefix(line, "\t") {
			break
		}
		if !inPgSQL || !strings.Contains(line, ":") {
			continue
		}
		parts := strings.SplitN(strings.TrimSpace(line), ":", 2)
		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
		value = strings.Trim(value, "\"")
		values[key] = value
	}
	cfg := dbConfig{
		Host:     firstNonEmpty(os.Getenv("GVA_PG_HOST"), values["path"]),
		Port:     firstNonEmpty(os.Getenv("GVA_PG_PORT"), values["port"], "5432"),
		Database: firstNonEmpty(os.Getenv("GVA_PG_DB"), values["db-name"]),
		Username: firstNonEmpty(os.Getenv("GVA_PG_USER"), values["username"]),
		Password: firstNonEmpty(os.Getenv("GVA_PG_PASSWORD"), values["password"]),
		Config:   values["config"],
	}
	if cfg.Host == "" || cfg.Port == "" || cfg.Database == "" || cfg.Username == "" {
		return dbConfig{}, fmt.Errorf("pgsql config is incomplete: host/port/database/username required")
	}
	return cfg, nil
}

func (c dbConfig) DSN() string {
	q := url.Values{}
	q.Set("sslmode", "disable")
	q.Set("timezone", "Asia/Shanghai")
	for _, field := range strings.Fields(c.Config) {
		kv := strings.SplitN(field, "=", 2)
		if len(kv) == 2 {
			key := strings.ToLower(strings.TrimSpace(kv[0]))
			value := strings.TrimSpace(kv[1])
			if strings.EqualFold(key, "timezone") {
				key = "timezone"
			}
			q.Set(key, value)
		}
	}
	u := url.URL{
		Scheme: "postgres",
		User:   url.UserPassword(c.Username, c.Password),
		Host:   net.JoinHostPort(c.Host, c.Port),
		Path:   "/" + c.Database,
	}
	u.RawQuery = q.Encode()
	return u.String()
}

func ensureSchema(ctx context.Context, db *sql.DB) error {
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS asset_categories (
			id bigserial PRIMARY KEY,
			created_at timestamptz,
			updated_at timestamptz,
			deleted_at timestamptz,
			name varchar(100) NOT NULL,
			code varchar(50) NOT NULL,
			description varchar(500),
			color varchar(20) DEFAULT '#334155',
			sort bigint DEFAULT 0,
			enabled boolean DEFAULT true
		)`,
		`CREATE UNIQUE INDEX IF NOT EXISTS idx_asset_categories_name ON asset_categories (name)`,
		`CREATE UNIQUE INDEX IF NOT EXISTS idx_asset_categories_code ON asset_categories (code)`,
		`CREATE INDEX IF NOT EXISTS idx_asset_categories_deleted_at ON asset_categories (deleted_at)`,
		`CREATE TABLE IF NOT EXISTS assets (
			id bigserial PRIMARY KEY,
			created_at timestamptz,
			updated_at timestamptz,
			deleted_at timestamptz,
			asset_code varchar(80) NOT NULL,
			name varchar(150) NOT NULL,
			category_id bigint NOT NULL,
			brand varchar(100),
			model varchar(120),
			serial_number varchar(120),
			quantity bigint NOT NULL DEFAULT 1,
			unit varchar(30) DEFAULT '件',
			unit_price numeric(16,2) NOT NULL DEFAULT 0,
			original_value numeric(18,2) NOT NULL DEFAULT 0,
			current_value numeric(18,2) NOT NULL DEFAULT 0,
			status varchar(30) NOT NULL DEFAULT 'pending_inbound',
			location varchar(150),
			custodian varchar(100),
			supplier varchar(150),
			purchase_date date,
			warranty_end_date date,
			photos jsonb,
			remarks text
		)`,
		`CREATE UNIQUE INDEX IF NOT EXISTS idx_assets_asset_code ON assets (asset_code)`,
		`CREATE INDEX IF NOT EXISTS idx_assets_name ON assets (name)`,
		`CREATE INDEX IF NOT EXISTS idx_assets_category_id ON assets (category_id)`,
		`CREATE INDEX IF NOT EXISTS idx_assets_status ON assets (status)`,
		`CREATE INDEX IF NOT EXISTS idx_assets_location ON assets (location)`,
		`CREATE INDEX IF NOT EXISTS idx_assets_custodian ON assets (custodian)`,
		`CREATE INDEX IF NOT EXISTS idx_assets_serial_number ON assets (serial_number)`,
		`CREATE INDEX IF NOT EXISTS idx_assets_deleted_at ON assets (deleted_at)`,
		`ALTER TABLE assets ALTER COLUMN status SET DEFAULT 'pending_inbound'`,
		`CREATE TABLE IF NOT EXISTS asset_operation_orders (
			id bigserial PRIMARY KEY,
			created_at timestamptz,
			updated_at timestamptz,
			deleted_at timestamptz,
			order_no varchar(80) NOT NULL,
			type varchar(30) NOT NULL,
			status varchar(20) NOT NULL DEFAULT 'draft',
			business_date date NOT NULL,
			target_location varchar(150),
			target_custodian varchar(100),
			reason varchar(500),
			remarks text,
			created_by bigint,
			created_by_name varchar(100),
			completed_by bigint,
			completed_by_name varchar(100),
			completed_at timestamptz
		)`,
		`CREATE UNIQUE INDEX IF NOT EXISTS idx_asset_operation_orders_order_no ON asset_operation_orders (order_no)`,
		`CREATE INDEX IF NOT EXISTS idx_asset_operation_orders_type ON asset_operation_orders (type)`,
		`CREATE INDEX IF NOT EXISTS idx_asset_operation_orders_status ON asset_operation_orders (status)`,
		`CREATE INDEX IF NOT EXISTS idx_asset_operation_orders_business_date ON asset_operation_orders (business_date)`,
		`CREATE INDEX IF NOT EXISTS idx_asset_operation_orders_deleted_at ON asset_operation_orders (deleted_at)`,
		`CREATE TABLE IF NOT EXISTS asset_operation_items (
			id bigserial PRIMARY KEY,
			created_at timestamptz,
			updated_at timestamptz,
			deleted_at timestamptz,
			order_id bigint NOT NULL,
			asset_id bigint NOT NULL,
			quantity bigint NOT NULL,
			asset_code varchar(80) NOT NULL,
			asset_name varchar(150) NOT NULL,
			from_status varchar(30),
			to_status varchar(30),
			from_location varchar(150),
			to_location varchar(150),
			from_custodian varchar(100),
			to_custodian varchar(100)
		)`,
		`CREATE INDEX IF NOT EXISTS idx_asset_operation_items_order_id ON asset_operation_items (order_id)`,
		`CREATE INDEX IF NOT EXISTS idx_asset_operation_items_asset_id ON asset_operation_items (asset_id)`,
		`CREATE INDEX IF NOT EXISTS idx_asset_operation_items_deleted_at ON asset_operation_items (deleted_at)`,
		`CREATE TABLE IF NOT EXISTS asset_operation_records (
			id bigserial PRIMARY KEY,
			created_at timestamptz,
			updated_at timestamptz,
			deleted_at timestamptz,
			order_id bigint NOT NULL,
			order_no varchar(80) NOT NULL,
			type varchar(30) NOT NULL,
			asset_id bigint NOT NULL,
			asset_code varchar(80) NOT NULL,
			asset_name varchar(150) NOT NULL,
			quantity bigint NOT NULL,
			from_status varchar(30),
			to_status varchar(30),
			from_location varchar(150),
			to_location varchar(150),
			from_custodian varchar(100),
			to_custodian varchar(100),
			operator_id bigint,
			operator_name varchar(100),
			operated_at timestamptz NOT NULL
		)`,
		`CREATE INDEX IF NOT EXISTS idx_asset_operation_records_order_id ON asset_operation_records (order_id)`,
		`CREATE INDEX IF NOT EXISTS idx_asset_operation_records_asset_id ON asset_operation_records (asset_id)`,
		`CREATE INDEX IF NOT EXISTS idx_asset_operation_records_type ON asset_operation_records (type)`,
		`CREATE INDEX IF NOT EXISTS idx_asset_operation_records_operated_at ON asset_operation_records (operated_at)`,
		`CREATE INDEX IF NOT EXISTS idx_asset_operation_records_deleted_at ON asset_operation_records (deleted_at)`,
	}
	for _, stmt := range stmts {
		if _, err := db.ExecContext(ctx, stmt); err != nil {
			return err
		}
	}
	return nil
}

func ensureDefaultCategories(ctx context.Context, db *sql.DB) error {
	for _, item := range defaultCategories {
		_, err := db.ExecContext(ctx, `
INSERT INTO asset_categories (created_at, updated_at, name, code, description, color, sort, enabled)
VALUES (NOW(), NOW(), $1, $2, $3, $4, $5, true)
ON CONFLICT (code) DO NOTHING`, item.Name, item.Code, item.Description, item.Color, item.Sort)
		if err != nil {
			return err
		}
	}
	return nil
}

func loadCategories(ctx context.Context, db *sql.DB) ([]category, error) {
	rows, err := db.QueryContext(ctx, `SELECT id, name, code FROM asset_categories WHERE deleted_at IS NULL AND enabled = true ORDER BY sort ASC, id ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []category
	for rows.Next() {
		var item category
		if err := rows.Scan(&item.ID, &item.Name, &item.Code); err != nil {
			return nil, err
		}
		list = append(list, item)
	}
	return list, rows.Err()
}

func pickTemplate(code string, index int) assetTemplate {
	items := templatesByCode[strings.ToUpper(code)]
	if len(items) == 0 {
		items = genericTemplates
	}
	return items[index%len(items)]
}

func statusForIndex(index int) string {
	pattern := []string{"pending_inbound", "idle", "idle", "in_use", "in_use", "in_use", "in_use", "in_use", "maintenance", "retired"}
	return pattern[index%len(pattern)]
}

func lifecycleAssignment(tpl assetTemplate, status string, index int) (string, string) {
	warehouses := []string{"资产仓库 A 区", "资产仓库 B 区", "信息设备备件库", "行政物资库"}
	switch status {
	case "pending_inbound":
		return "", ""
	case "idle":
		return warehouses[index%len(warehouses)], ""
	case "maintenance":
		locations := []string{"设备维修中心", "外送维修待验区", "信息设备检修间"}
		return locations[index%len(locations)], "设备保障组-顾维"
	case "retired":
		return "报废暂存区", ""
	default:
		return tpl.Location, tpl.Custodian
	}
}

func lifecycleRemark(categoryName, status, location string) string {
	remarks := map[string]string{
		"pending_inbound": "采购验收资料已齐全，等待办理正式入库。",
		"idle":            "已完成入库，当前处于库存可领用状态。",
		"in_use":          "已完成领用交接，由当前责任部门日常保管。",
		"maintenance":     "设备故障已登记，正在维修并等待检测结果。",
		"retired":         "已完成报废处置，资产估值归零并保留审计记录。",
	}
	result := fmt.Sprintf("%s演示档案：%s", categoryName, remarks[status])
	if location != "" {
		result += " 当前位置：" + location + "。"
	}
	return result
}

type operationSnapshot struct {
	OperationType  string
	FromStatus     string
	ToStatus       string
	FromLocation   string
	ToLocation     string
	FromCustodian  string
	ToCustodian    string
	TargetLocation string
	TargetKeeper   string
	Reason         string
}

type seededDraft struct {
	Asset    seededAsset
	Snapshot operationSnapshot
}

var draftOperationTypes = []string{"inbound", "issue", "transfer", "return", "maintenance", "scrap"}

func draftAssetEligible(operationType, status string) bool {
	allowed := map[string]map[string]struct{}{
		"inbound":     {"pending_inbound": {}},
		"issue":       {"idle": {}},
		"transfer":    {"idle": {}, "in_use": {}},
		"return":      {"in_use": {}, "maintenance": {}},
		"maintenance": {"idle": {}, "in_use": {}},
		"scrap":       {"idle": {}, "in_use": {}, "maintenance": {}},
	}
	_, ok := allowed[operationType][status]
	return ok
}

func draftSnapshotFor(operationType string, asset seededAsset) operationSnapshot {
	snapshot := operationSnapshot{
		OperationType: operationType,
		FromStatus:    asset.Status, FromLocation: asset.Location, FromCustodian: asset.Custodian,
	}
	switch operationType {
	case "inbound":
		snapshot.ToStatus = "idle"
		snapshot.ToLocation = "资产仓库 C 区"
		snapshot.TargetLocation = snapshot.ToLocation
		snapshot.Reason = "采购资产验收入库"
	case "issue":
		snapshot.ToStatus = "in_use"
		snapshot.ToLocation = "项目研发区"
		snapshot.ToCustodian = "研发部-张明"
		snapshot.TargetLocation = snapshot.ToLocation
		snapshot.TargetKeeper = snapshot.ToCustodian
		snapshot.Reason = "项目岗位资产领用"
	case "transfer":
		snapshot.ToStatus = asset.Status
		snapshot.ToLocation = "二号办公楼 3F"
		snapshot.ToCustodian = asset.Custodian
		if snapshot.ToCustodian == "" {
			snapshot.ToCustodian = "行政部-李杰"
		}
		snapshot.TargetLocation = snapshot.ToLocation
		snapshot.TargetKeeper = snapshot.ToCustodian
		snapshot.Reason = "办公区域与责任人调整"
	case "return":
		snapshot.ToStatus = "idle"
		snapshot.ToLocation = "资产仓库 A 区"
		snapshot.TargetLocation = snapshot.ToLocation
		snapshot.Reason = "阶段工作结束归还资产"
	case "maintenance":
		snapshot.ToStatus = "maintenance"
		snapshot.ToLocation = "设备维修中心"
		snapshot.ToCustodian = "设备保障组-顾维"
		snapshot.TargetLocation = snapshot.ToLocation
		snapshot.TargetKeeper = snapshot.ToCustodian
		snapshot.Reason = "设备运行异常，申请检修"
	case "scrap":
		snapshot.ToStatus = "retired"
		snapshot.ToLocation = "报废暂存区"
		snapshot.TargetLocation = snapshot.ToLocation
		snapshot.Reason = "超过经济使用年限，申请报废"
	}
	return snapshot
}

func buildDraftSeeds(assets []seededAsset, perType int) ([]seededDraft, error) {
	if perType <= 0 {
		return nil, errors.New("draft count per operation must be greater than zero")
	}
	drafts := make([]seededDraft, 0, len(draftOperationTypes)*perType)
	usedAssets := map[int64]struct{}{}
	for _, operationType := range draftOperationTypes {
		selected := 0
		for _, asset := range assets {
			if _, used := usedAssets[asset.ID]; used || !draftAssetEligible(operationType, asset.Status) {
				continue
			}
			drafts = append(drafts, seededDraft{Asset: asset, Snapshot: draftSnapshotFor(operationType, asset)})
			usedAssets[asset.ID] = struct{}{}
			selected++
			if selected == perType {
				break
			}
		}
		if selected != perType {
			return nil, fmt.Errorf("insufficient eligible assets for %s drafts: got %d, want %d", operationType, selected, perType)
		}
	}
	return drafts, nil
}

func operationSnapshotFor(asset seededAsset, index int) operationSnapshot {
	warehouse := []string{"资产仓库 A 区", "资产仓库 B 区", "信息设备备件库", "行政物资库"}[index%4]
	switch asset.Status {
	case "idle":
		if index%3 == 0 {
			return operationSnapshot{
				OperationType: "return", FromStatus: "in_use", ToStatus: "idle",
				FromLocation: "业务部门使用区", ToLocation: asset.Location,
				FromCustodian: "业务部门-使用人", TargetLocation: asset.Location,
				Reason: "阶段工作结束，资产归还仓库统一保管",
			}
		}
		return operationSnapshot{
			OperationType: "inbound", FromStatus: "pending_inbound", ToStatus: "idle",
			ToLocation: asset.Location, TargetLocation: asset.Location,
			Reason: "采购验收合格，办理正式入库",
		}
	case "in_use":
		if index%4 == 0 {
			return operationSnapshot{
				OperationType: "transfer", FromStatus: "in_use", ToStatus: "in_use",
				FromLocation: "原使用部门", ToLocation: asset.Location,
				FromCustodian: "原责任部门-使用人", ToCustodian: asset.Custodian,
				TargetLocation: asset.Location, TargetKeeper: asset.Custodian,
				Reason: "项目和责任部门调整，办理资产调拨",
			}
		}
		return operationSnapshot{
			OperationType: "issue", FromStatus: "idle", ToStatus: "in_use",
			FromLocation: warehouse, ToLocation: asset.Location,
			ToCustodian: asset.Custodian, TargetLocation: asset.Location, TargetKeeper: asset.Custodian,
			Reason: "业务岗位配置，办理资产领用",
		}
	case "maintenance":
		return operationSnapshot{
			OperationType: "maintenance", FromStatus: "in_use", ToStatus: "maintenance",
			FromLocation: "业务部门使用区", ToLocation: asset.Location,
			FromCustodian: "业务部门-使用人", ToCustodian: asset.Custodian,
			TargetLocation: asset.Location, TargetKeeper: asset.Custodian,
			Reason: "运行异常，送修检测并安排维护",
		}
	case "retired":
		fromStatus := "idle"
		fromLocation := warehouse
		if index%2 == 0 {
			fromStatus = "maintenance"
			fromLocation = "设备维修中心"
		}
		return operationSnapshot{
			OperationType: "scrap", FromStatus: fromStatus, ToStatus: "retired",
			FromLocation: fromLocation, ToLocation: asset.Location,
			TargetLocation: asset.Location, Reason: "超过经济使用年限且维修价值较低，批准报废处置",
		}
	default:
		return operationSnapshot{}
	}
}

func seedOperationData(ctx context.Context, tx *sql.Tx, assets []seededAsset, prefix string, now time.Time) (map[string]int, error) {
	stats := map[string]int{}
	var operatorID int64 = 1
	operatorName := "系统管理员"
	_ = tx.QueryRowContext(ctx, `SELECT id, COALESCE(NULLIF(nick_name, ''), username) FROM sys_users WHERE username = 'admin' AND deleted_at IS NULL ORDER BY id LIMIT 1`).Scan(&operatorID, &operatorName)

	operationPrefix := "SEED-" + sanitizeCode(prefix)
	for index, asset := range assets {
		snapshot := operationSnapshotFor(asset, index)
		if snapshot.OperationType == "" {
			continue
		}
		operatedAt := asset.CreatedAt.Add(2 * time.Hour)
		if operatedAt.After(now) {
			operatedAt = now
		}
		orderNo := fmt.Sprintf("%s-%s-%04d", operationPrefix, operationCode(snapshot.OperationType), index+1)
		var orderID int64
		err := tx.QueryRowContext(ctx, `
INSERT INTO asset_operation_orders (
  created_at, updated_at, order_no, type, status, business_date,
  target_location, target_custodian, reason, remarks,
  created_by, created_by_name, completed_by, completed_by_name, completed_at
) VALUES ($1, $1, $2, $3, 'completed', $4, $5, $6, $7, $8, $9, $10, $9, $10, $1)
RETURNING id`, operatedAt, orderNo, snapshot.OperationType, operatedAt.Format("2006-01-02"),
			snapshot.TargetLocation, snapshot.TargetKeeper, snapshot.Reason,
			"系统演示数据：用于验证资产全生命周期单据、状态与审计快照。", operatorID, operatorName,
		).Scan(&orderID)
		if err != nil {
			return stats, err
		}

		_, err = tx.ExecContext(ctx, `
INSERT INTO asset_operation_items (
  created_at, updated_at, order_id, asset_id, quantity, asset_code, asset_name,
  from_status, to_status, from_location, to_location, from_custodian, to_custodian
) VALUES ($1, $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)`,
			operatedAt, orderID, asset.ID, asset.Quantity, asset.Code, asset.Name,
			snapshot.FromStatus, snapshot.ToStatus, snapshot.FromLocation, snapshot.ToLocation,
			snapshot.FromCustodian, snapshot.ToCustodian,
		)
		if err != nil {
			return stats, err
		}

		_, err = tx.ExecContext(ctx, `
INSERT INTO asset_operation_records (
  created_at, updated_at, order_id, order_no, type, asset_id, asset_code, asset_name, quantity,
  from_status, to_status, from_location, to_location, from_custodian, to_custodian,
  operator_id, operator_name, operated_at
) VALUES ($1, $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $1)`,
			operatedAt, orderID, orderNo, snapshot.OperationType, asset.ID, asset.Code, asset.Name, asset.Quantity,
			snapshot.FromStatus, snapshot.ToStatus, snapshot.FromLocation, snapshot.ToLocation,
			snapshot.FromCustodian, snapshot.ToCustodian, operatorID, operatorName,
		)
		if err != nil {
			return stats, err
		}
		stats[snapshot.OperationType]++
	}
	return stats, nil
}

func seedOperationDrafts(ctx context.Context, tx *sql.Tx, assets []seededAsset, prefix string, now time.Time, perType int) (map[string]int, error) {
	drafts, err := buildDraftSeeds(assets, perType)
	if err != nil {
		return nil, err
	}
	stats := map[string]int{}
	var operatorID int64 = 1
	operatorName := "系统管理员"
	_ = tx.QueryRowContext(ctx, `SELECT id, COALESCE(NULLIF(nick_name, ''), username) FROM sys_users WHERE username = 'admin' AND deleted_at IS NULL ORDER BY id LIMIT 1`).Scan(&operatorID, &operatorName)

	operationPrefix := "SEED-" + sanitizeCode(prefix)
	for index, draft := range drafts {
		snapshot := draft.Snapshot
		asset := draft.Asset
		sequence := stats[snapshot.OperationType] + 1
		createdAt := now.Add(-time.Duration(len(drafts)-index) * time.Minute)
		orderNo := fmt.Sprintf("%s-DRAFT-%s-%02d", operationPrefix, operationCode(snapshot.OperationType), sequence)
		var orderID int64
		err := tx.QueryRowContext(ctx, `
INSERT INTO asset_operation_orders (
  created_at, updated_at, order_no, type, status, business_date,
  target_location, target_custodian, reason, remarks, created_by, created_by_name
) VALUES ($1, $1, $2, $3, 'draft', $4, $5, $6, $7, $8, $9, $10)
RETURNING id`, createdAt, orderNo, snapshot.OperationType, createdAt.Format("2006-01-02"),
			snapshot.TargetLocation, snapshot.TargetKeeper, snapshot.Reason,
			"可编辑演示草稿：修改后可保存、提交或删除。", operatorID, operatorName,
		).Scan(&orderID)
		if err != nil {
			return stats, err
		}
		_, err = tx.ExecContext(ctx, `
INSERT INTO asset_operation_items (
  created_at, updated_at, order_id, asset_id, quantity, asset_code, asset_name,
  from_status, to_status, from_location, to_location, from_custodian, to_custodian
) VALUES ($1, $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)`,
			createdAt, orderID, asset.ID, asset.Quantity, asset.Code, asset.Name,
			snapshot.FromStatus, snapshot.ToStatus, snapshot.FromLocation, snapshot.ToLocation,
			snapshot.FromCustodian, snapshot.ToCustodian,
		)
		if err != nil {
			return stats, err
		}
		stats[snapshot.OperationType]++
	}
	return stats, nil
}

func operationCode(operationType string) string {
	codes := map[string]string{"inbound": "RK", "issue": "LY", "transfer": "DB", "return": "GH", "maintenance": "WX", "scrap": "BF"}
	return codes[operationType]
}

func operationLabel(operationType string) string {
	labels := map[string]string{"inbound": "入库", "issue": "领用", "transfer": "调拨", "return": "归还", "maintenance": "维修", "scrap": "报废"}
	return labels[operationType]
}

func depreciatedValue(original float64, purchaseDate time.Time, status string, rng *rand.Rand) float64 {
	if original <= 0 {
		return 0
	}
	months := math.Max(1, time.Since(purchaseDate).Hours()/24/30)
	depreciation := 0.05 + months*0.012 + rng.Float64()*0.10
	if status == "idle" {
		depreciation += 0.06
	}
	if status == "maintenance" {
		depreciation += 0.14
	}
	if status == "retired" {
		return 0
	}
	if depreciation > 0.82 {
		depreciation = 0.82
	}
	return money(original * (1 - depreciation))
}

func money(value float64) float64 {
	return math.Round(value*100) / 100
}

func sanitizeCode(code string) string {
	re := regexp.MustCompile(`[^A-Z0-9]+`)
	return strings.Trim(re.ReplaceAllString(strings.ToUpper(code), ""), "-")
}

func statusLabel(status string) string {
	switch status {
	case "pending_inbound":
		return "待入库"
	case "in_use":
		return "使用中"
	case "idle":
		return "闲置"
	case "maintenance":
		return "维修中"
	case "retired":
		return "已处置"
	default:
		return status
	}
}

func sortedKeys(m map[string]int) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}
