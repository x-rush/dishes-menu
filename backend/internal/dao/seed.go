package dao

import (
	"context"
	"fmt"

	"dishes-menu/internal/model"
)

// SeedDefaultDishes inserts 30 default Chinese home dishes if the dishes table is empty.
// Returns the number of dishes actually inserted (0 if the table already had data).
func (r *DishRepo) SeedDefaultDishes(ctx context.Context) (int, error) {
	n, err := r.Count(ctx)
	if err != nil {
		return 0, err
	}
	if n > 0 {
		return 0, nil
	}
	for _, d := range defaultSeedDishes {
		if err := r.Create(ctx, d); err != nil {
			return 0, fmt.Errorf("seed %s: %w", d.Name, err)
		}
	}
	return len(defaultSeedDishes), nil
}

// defaultSeedDishes: 30 道家常菜，覆盖早/午/晚/加餐
var defaultSeedDishes = []*model.Dish{
	mkDish("白粥配咸蛋", []model.Slot{model.SlotBreakfast, model.SlotDinner}, 30, []string{"大米", "咸鸭蛋"}, "白粥煮到开花，咸蛋切瓣摆盘。"),
	mkDish("煎蛋三明治", []model.Slot{model.SlotBreakfast, model.SlotSnack}, 10, []string{"全麦面包", "鸡蛋", "生菜", "番茄"}, "全麦面包+煎蛋+生菜+番茄沙司。"),
	mkDish("葱花饼", []model.Slot{model.SlotBreakfast, model.SlotSnack}, 25, []string{"面粉", "葱"}, "面团抹油撒葱花卷起擀平，两面煎金黄。"),
	mkDish("小米南瓜粥", []model.Slot{model.SlotBreakfast, model.SlotDinner}, 30, []string{"小米", "南瓜", "红枣"}, "小米+南瓜块煮 30 分钟，可加红枣。"),
	mkDish("番茄鸡蛋面", []model.Slot{model.SlotBreakfast, model.SlotLunch, model.SlotDinner}, 15, []string{"挂面", "番茄", "鸡蛋", "葱花"}, "番茄切块炒出汁，加水煮开下挂面，鸡蛋打散淋入。"),
	mkDish("豆浆油条", []model.Slot{model.SlotBreakfast, model.SlotSnack}, 15, []string{"黄豆", "油条"}, "豆浆提前泡豆预约，油条复炸 30 秒。"),
	mkDish("蛋炒饭", []model.Slot{model.SlotBreakfast, model.SlotLunch, model.SlotDinner}, 10, []string{"米饭", "鸡蛋", "葱花", "酱油"}, "隔夜饭+鸡蛋+葱花+酱油大火快炒。"),
	mkDish("燕麦牛奶碗", []model.Slot{model.SlotBreakfast, model.SlotSnack}, 5, []string{"燕麦", "牛奶", "蓝莓", "香蕉"}, "燕麦+热牛奶泡 3 分钟，加蓝莓香蕉。"),
	mkDish("肉包子", []model.Slot{model.SlotBreakfast, model.SlotSnack}, 60, []string{"面粉", "猪肉馅", "葱", "姜", "生抽"}, "猪肉馅+葱姜+生抽，擀皮包 18 个褶。"),
	mkDish("八宝粥", []model.Slot{model.SlotBreakfast, model.SlotDinner}, 50, []string{"红豆", "花生", "红枣", "桂圆", "莲子", "糯米", "薏米", "百合"}, "红豆+花生+红枣+桂圆+莲子+糯米+薏米+百合，电压力锅 40 分钟。"),
	mkDish("番茄炒蛋", []model.Slot{model.SlotLunch, model.SlotDinner}, 10, []string{"番茄", "鸡蛋", "葱花"}, "鸡蛋先炒定型盛出，番茄炒出汁再合。"),
	mkDish("蒜蓉西兰花", []model.Slot{model.SlotLunch, model.SlotDinner}, 10, []string{"西兰花", "大蒜"}, "西兰花焯水 1 分钟过凉水，蒜末爆香快炒。"),
	mkDish("清炒时蔬", []model.Slot{model.SlotLunch, model.SlotDinner}, 8, []string{"油菜", "大蒜"}, "油菜/上海青焯水后蒜末炒。"),
	mkDish("麻婆豆腐", []model.Slot{model.SlotLunch, model.SlotDinner}, 25, []string{"嫩豆腐", "豆瓣酱", "花椒粉", "肉末"}, "嫩豆腐切块盐水泡 10 分钟，豆瓣酱+花椒粉+肉末。"),
	mkDish("鱼香肉丝", []model.Slot{model.SlotLunch, model.SlotDinner}, 25, []string{"里脊肉", "木耳", "胡萝卜", "醋", "糖", "酱油", "料酒"}, "里脊丝+木耳+胡萝卜丝，鱼香汁（醋糖酱油料酒）。"),
	mkDish("宫保鸡丁", []model.Slot{model.SlotLunch, model.SlotDinner}, 30, []string{"鸡腿肉", "花生米", "干辣椒", "花椒"}, "鸡腿肉丁+花生米+干辣椒+花椒，糖醋汁。"),
	mkDish("红烧肉", []model.Slot{model.SlotLunch, model.SlotDinner}, 60, []string{"五花肉", "冰糖", "料酒", "老抽"}, "五花肉切块焯水，冰糖炒糖色，料酒老抽慢炖 40 分钟。"),
	mkDish("糖醋排骨", []model.Slot{model.SlotLunch, model.SlotDinner}, 45, []string{"排骨", "糖", "醋"}, "排骨焯水，煎至金黄，糖醋汁收汁。"),
	mkDish("清蒸鲈鱼", []model.Slot{model.SlotLunch, model.SlotDinner}, 20, []string{"鲈鱼", "葱", "姜", "蒸鱼豉油"}, "鲈鱼葱姜铺底蒸 8 分钟，泼热油+蒸鱼豉油。"),
	mkDish("香菇滑鸡", []model.Slot{model.SlotLunch, model.SlotDinner}, 30, []string{"鸡腿肉", "香菇", "姜", "生粉"}, "鸡腿肉+香菇+姜丝+生粉，蒸 20 分钟。"),
	mkDish("紫菜蛋花汤", []model.Slot{model.SlotLunch, model.SlotDinner}, 8, []string{"紫菜", "虾皮", "鸡蛋", "香油"}, "水烧开+紫菜+虾皮+蛋液+香油+盐。"),
	mkDish("番茄牛腩汤", []model.Slot{model.SlotLunch, model.SlotDinner}, 90, []string{"牛腩", "番茄", "洋葱", "姜"}, "牛腩焯水+番茄+洋葱+姜炖 1.5 小时。"),
	mkDish("冬瓜排骨汤", []model.Slot{model.SlotLunch, model.SlotDinner}, 60, []string{"排骨", "冬瓜", "姜"}, "排骨+冬瓜+姜片炖 1 小时。"),
	mkDish("银耳莲子羹", []model.Slot{model.SlotSnack, model.SlotDinner}, 50, []string{"银耳", "莲子", "红枣", "冰糖"}, "银耳+莲子+红枣+冰糖煮 40 分钟。"),
	mkDish("酸奶水果杯", []model.Slot{model.SlotSnack, model.SlotBreakfast}, 5, []string{"希腊酸奶", "草莓", "蓝莓", "燕麦"}, "希腊酸奶+草莓+蓝莓+燕麦。"),
	mkDish("卤鸡爪", []model.Slot{model.SlotSnack, model.SlotDinner}, 40, []string{"鸡爪", "八角", "桂皮", "老抽", "生抽", "冰糖"}, "鸡爪焯水+八角+桂皮+老抽+生抽+冰糖卤 30 分钟。"),
	mkDish("卤蛋", []model.Slot{model.SlotSnack}, 30, []string{"鸡蛋", "卤汁"}, "白水蛋剥壳泡卤汁 4 小时。"),
	mkDish("烤红薯", []model.Slot{model.SlotSnack, model.SlotBreakfast}, 60, []string{"红薯"}, "红薯 200° 烤箱 60 分钟，或微波炉高火 8 分钟。"),
	mkDish("蒸南瓜", []model.Slot{model.SlotSnack, model.SlotBreakfast, model.SlotDinner}, 25, []string{"南瓜", "蜂蜜"}, "南瓜切块蒸 20 分钟，淋蜂蜜。"),
	mkDish("红枣桂圆茶", []model.Slot{model.SlotSnack, model.SlotBreakfast}, 15, []string{"红枣", "桂圆", "枸杞", "冰糖"}, "红枣+桂圆+枸杞+冰糖煮 10 分钟。"),
}

// mkDish builds a Dish with the simplified schema.
func mkDish(name string, slots []model.Slot, estimatedTime int, ingredients []string, note string) *model.Dish {
	return &model.Dish{
		Name:          name,
		Slots:         slots,
		Ingredients:   ingredients,
		EstimatedTime: estimatedTime,
		Note:          note,
	}
}
