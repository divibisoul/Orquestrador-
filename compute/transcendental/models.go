package transcendental

type Model struct { Name string; Vendor string; Family string; Supports []string; MemoryGB float64 }

var Models = []Model{
	{Name:"blackwell",Vendor:"nvidia",Family:"Blackwell",Supports:[]string{"fp32","fp64"},MemoryGB:192},
	{Name:"vera_rubin",Vendor:"nvidia",Family:"Vera Rubin",Supports:[]string{"fp32","fp64"},MemoryGB:288},
	{Name:"mi400",Vendor:"amd",Family:"MI400",Supports:[]string{"fp32","fp64"},MemoryGB:288},
	{Name:"atlas",Vendor:"amd",Family:"Atlas",Supports:[]string{"fp32","fp64"},MemoryGB:256},
	{Name:"trillium",Vendor:"google",Family:"Trillium",Supports:[]string{"fp32"},MemoryGB:32},
}
