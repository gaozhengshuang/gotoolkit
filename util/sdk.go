package util
import "sort"
import "gitee.com/jntse/gotoolkit/log"

// --------------------------------------------------------------------------
/// @brief 返回指定key的排序组合串
/// @brief key默认升序排序
/// @brief dict 原词典
/// @brief keys 需要排序的keys
/// @brief sep 分隔符
/// @return 
// --------------------------------------------------------------------------
func SpliceSortKeyValue(dict map[string]interface{}, keys []string, sep string) string {
	vdict := MakeLiteVarTypeStringMap(dict)

	sortkeys := make(sort.StringSlice, 0)
	sortkeys = append(sortkeys, keys...)
	sortkeys.Sort()
	splicestr := ""

	for i, v := range sortkeys {
		val, find := vdict[v]
		if find == false {
			log.Error("[GetSortKeyValue] vdict not have key[%s]", v)
			return ""		// fatal error
		}

		splicestr += v
		splicestr += "="
		splicestr += val.String()
		if i < len(sortkeys) - 1 {
			splicestr += sep
		}
	}
	return splicestr
}


