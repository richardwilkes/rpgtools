// Copyright ©2017-2022 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package namer

import "github.com/richardwilkes/rpgtools/names"

// Male is a random name generator for male first names. Created from data
// obtained from http://www.ssa.gov/oact/babynames/names.zip
var Male = (&names.Data{
	StartsWithVowelFreq:     7008,
	StartsWithConsonantFreq: 32720,
	CountFreq:               []int{0, 812, 3921, 8842, 17235, 4949, 2953, 616, 270, 71, 31, 10, 17, 1},
	Segments: [names.ArraySize][]names.Segment{
		{
			{Value: "a", Freq: 3028},
			{Value: "aa", Freq: 132},
			{Value: "ae", Freq: 27},
			{Value: "ai", Freq: 84},
			{Value: "au", Freq: 128},
			{Value: "ay", Freq: 51},
			{Value: "aya", Freq: 8},
			{Value: "ayo", Freq: 9},
			{Value: "ayu", Freq: 5},
			{Value: "e", Freq: 1308},
			{Value: "ea", Freq: 67},
			{Value: "ei", Freq: 41},
			{Value: "eu", Freq: 59},
			{Value: "ey", Freq: 10},
			{Value: "eya", Freq: 5},
			{Value: "i", Freq: 520},
			{Value: "ia", Freq: 12},
			{Value: "io", Freq: 10},
			{Value: "o", Freq: 685},
			{Value: "oa", Freq: 17},
			{Value: "ou", Freq: 5},
			{Value: "u", Freq: 144},
			{Value: "y", Freq: 24},
			{Value: "ya", Freq: 192},
			{Value: "yai", Freq: 6},
			{Value: "ye", Freq: 59},
			{Value: "yei", Freq: 6},
			{Value: "yi", Freq: 29},
			{Value: "yo", Freq: 112},
			{Value: "yoa", Freq: 8},
			{Value: "you", Freq: 20},
			{Value: "yu", Freq: 64},
		},
		{
			{Value: "a", Freq: 18780},
			{Value: "aa", Freq: 330},
			{Value: "aai", Freq: 6},
			{Value: "ae", Freq: 563},
			{Value: "aee", Freq: 8},
			{Value: "aeu", Freq: 7},
			{Value: "ai", Freq: 1048},
			{Value: "aia", Freq: 57},
			{Value: "aie", Freq: 11},
			{Value: "aii", Freq: 6},
			{Value: "aio", Freq: 7},
			{Value: "aiy", Freq: 6},
			{Value: "aiya", Freq: 8},
			{Value: "aiyo", Freq: 5},
			{Value: "ao", Freq: 45},
			{Value: "au", Freq: 574},
			{Value: "aue", Freq: 6},
			{Value: "ay", Freq: 1087},
			{Value: "aya", Freq: 52},
			{Value: "ayaa", Freq: 5},
			{Value: "aye", Freq: 39},
			{Value: "ayi", Freq: 13},
			{Value: "ayio", Freq: 5},
			{Value: "ayo", Freq: 27},
			{Value: "ayu", Freq: 6},
			{Value: "ayya", Freq: 6},
			{Value: "e", Freq: 12813},
			{Value: "ea", Freq: 690},
			{Value: "eai", Freq: 5},
			{Value: "eau", Freq: 28},
			{Value: "ee", Freq: 725},
			{Value: "eea", Freq: 7},
			{Value: "ei", Freq: 436},
			{Value: "eia", Freq: 11},
			{Value: "eio", Freq: 46},
			{Value: "eo", Freq: 429},
			{Value: "eou", Freq: 6},
			{Value: "eu", Freq: 79},
			{Value: "ey", Freq: 262},
			{Value: "eya", Freq: 23},
			{Value: "eye", Freq: 16},
			{Value: "eyo", Freq: 43},
			{Value: "eyu", Freq: 6},
			{Value: "i", Freq: 8858},
			{Value: "ia", Freq: 1066},
			{Value: "iaa", Freq: 8},
			{Value: "iai", Freq: 6},
			{Value: "ie", Freq: 999},
			{Value: "iea", Freq: 7},
			{Value: "ieo", Freq: 16},
			{Value: "ieu", Freq: 5},
			{Value: "ii", Freq: 21},
			{Value: "io", Freq: 725},
			{Value: "iou", Freq: 193},
			{Value: "iu", Freq: 362},
			{Value: "iy", Freq: 12},
			{Value: "iya", Freq: 83},
			{Value: "iyaa", Freq: 8},
			{Value: "iye", Freq: 7},
			{Value: "iyo", Freq: 40},
			{Value: "o", Freq: 8831},
			{Value: "oa", Freq: 113},
			{Value: "oao", Freq: 6},
			{Value: "oe", Freq: 76},
			{Value: "oi", Freq: 71},
			{Value: "oo", Freq: 150},
			{Value: "ou", Freq: 231},
			{Value: "oua", Freq: 5},
			{Value: "oui", Freq: 6},
			{Value: "oy", Freq: 71},
			{Value: "oya", Freq: 17},
			{Value: "oye", Freq: 14},
			{Value: "u", Freq: 2591},
			{Value: "ua", Freq: 570},
			{Value: "uaa", Freq: 6},
			{Value: "uai", Freq: 11},
			{Value: "uay", Freq: 16},
			{Value: "ue", Freq: 265},
			{Value: "uee", Freq: 7},
			{Value: "ui", Freq: 299},
			{Value: "uia", Freq: 7},
			{Value: "uie", Freq: 19},
			{Value: "uo", Freq: 46},
			{Value: "uu", Freq: 6},
			{Value: "uy", Freq: 8},
			{Value: "uyo", Freq: 7},
			{Value: "y", Freq: 2211},
			{Value: "ya", Freq: 231},
			{Value: "yaa", Freq: 11},
			{Value: "yai", Freq: 19},
			{Value: "ye", Freq: 116},
			{Value: "yi", Freq: 22},
			{Value: "yia", Freq: 13},
			{Value: "yie", Freq: 10},
			{Value: "yio", Freq: 15},
			{Value: "yo", Freq: 100},
			{Value: "yu", Freq: 37},
		},
		{
			{Value: "a", Freq: 1390},
			{Value: "aa", Freq: 8},
			{Value: "ae", Freq: 182},
			{Value: "ai", Freq: 207},
			{Value: "aia", Freq: 10},
			{Value: "aio", Freq: 5},
			{Value: "ao", Freq: 36},
			{Value: "au", Freq: 11},
			{Value: "ay", Freq: 282},
			{Value: "aya", Freq: 16},
			{Value: "aye", Freq: 22},
			{Value: "ayo", Freq: 13},
			{Value: "e", Freq: 3073},
			{Value: "ea", Freq: 60},
			{Value: "eau", Freq: 8},
			{Value: "ee", Freq: 313},
			{Value: "ei", Freq: 52},
			{Value: "eo", Freq: 68},
			{Value: "eu", Freq: 6},
			{Value: "ey", Freq: 657},
			{Value: "eya", Freq: 6},
			{Value: "eyo", Freq: 6},
			{Value: "i", Freq: 1428},
			{Value: "ia", Freq: 187},
			{Value: "ie", Freq: 787},
			{Value: "ieu", Freq: 5},
			{Value: "ii", Freq: 27},
			{Value: "io", Freq: 433},
			{Value: "iy", Freq: 18},
			{Value: "iya", Freq: 20},
			{Value: "o", Freq: 1560},
			{Value: "oa", Freq: 25},
			{Value: "oe", Freq: 30},
			{Value: "oey", Freq: 6},
			{Value: "oi", Freq: 12},
			{Value: "oo", Freq: 13},
			{Value: "ou", Freq: 36},
			{Value: "oua", Freq: 9},
			{Value: "oy", Freq: 69},
			{Value: "oya", Freq: 5},
			{Value: "oye", Freq: 11},
			{Value: "u", Freq: 178},
			{Value: "ua", Freq: 39},
			{Value: "uay", Freq: 7},
			{Value: "ue", Freq: 114},
			{Value: "ui", Freq: 11},
			{Value: "uie", Freq: 5},
			{Value: "uio", Freq: 6},
			{Value: "uo", Freq: 18},
			{Value: "uy", Freq: 8},
			{Value: "y", Freq: 1201},
			{Value: "ya", Freq: 73},
			{Value: "yae", Freq: 7},
			{Value: "ye", Freq: 66},
			{Value: "yi", Freq: 9},
			{Value: "yo", Freq: 5},
			{Value: "yu", Freq: 6},
		},
		{
			{Value: "b", Freq: 908},
			{Value: "bh", Freq: 9},
			{Value: "bl", Freq: 75},
			{Value: "br", Freq: 669},
			{Value: "c", Freq: 1320},
			{Value: "ch", Freq: 452},
			{Value: "chr", Freq: 118},
			{Value: "cl", Freq: 219},
			{Value: "cr", Freq: 125},
			{Value: "d", Freq: 3065},
			{Value: "dh", Freq: 29},
			{Value: "dhr", Freq: 5},
			{Value: "dj", Freq: 11},
			{Value: "dm", Freq: 30},
			{Value: "dr", Freq: 163},
			{Value: "dv", Freq: 8},
			{Value: "dw", Freq: 20},
			{Value: "f", Freq: 396},
			{Value: "fl", Freq: 56},
			{Value: "fr", Freq: 130},
			{Value: "g", Freq: 819},
			{Value: "gh", Freq: 12},
			{Value: "gl", Freq: 56},
			{Value: "gr", Freq: 182},
			{Value: "gw", Freq: 11},
			{Value: "h", Freq: 1086},
			{Value: "hr", Freq: 16},
			{Value: "j", Freq: 3797},
			{Value: "jc", Freq: 7},
			{Value: "jd", Freq: 7},
			{Value: "jh", Freq: 68},
			{Value: "jl", Freq: 7},
			{Value: "jm", Freq: 9},
			{Value: "jr", Freq: 8},
			{Value: "js", Freq: 7},
			{Value: "jsh", Freq: 5},
			{Value: "jv", Freq: 12},
			{Value: "jw", Freq: 6},
			{Value: "k", Freq: 2580},
			{Value: "kh", Freq: 213},
			{Value: "khr", Freq: 8},
			{Value: "kj", Freq: 5},
			{Value: "kl", Freq: 16},
			{Value: "kn", Freq: 19},
			{Value: "kr", Freq: 127},
			{Value: "kv", Freq: 5},
			{Value: "kw", Freq: 35},
			{Value: "l", Freq: 1762},
			{Value: "ll", Freq: 12},
			{Value: "m", Freq: 2517},
			{Value: "mc", Freq: 11},
			{Value: "mcc", Freq: 13},
			{Value: "mccl", Freq: 5},
			{Value: "mck", Freq: 32},
			{Value: "n", Freq: 1136},
			{Value: "ng", Freq: 7},
			{Value: "nh", Freq: 5},
			{Value: "p", Freq: 404},
			{Value: "ph", Freq: 80},
			{Value: "pl", Freq: 18},
			{Value: "pr", Freq: 150},
			{Value: "q", Freq: 288},
			{Value: "qw", Freq: 6},
			{Value: "r", Freq: 2010},
			{Value: "rh", Freq: 64},
			{Value: "s", Freq: 1325},
			{Value: "sc", Freq: 17},
			{Value: "sch", Freq: 15},
			{Value: "sh", Freq: 695},
			{Value: "shl", Freq: 11},
			{Value: "shm", Freq: 5},
			{Value: "shr", Freq: 27},
			{Value: "sk", Freq: 32},
			{Value: "sl", Freq: 20},
			{Value: "sm", Freq: 11},
			{Value: "sn", Freq: 11},
			{Value: "sp", Freq: 29},
			{Value: "sr", Freq: 36},
			{Value: "st", Freq: 204},
			{Value: "str", Freq: 15},
			{Value: "sw", Freq: 17},
			{Value: "t", Freq: 1750},
			{Value: "th", Freq: 230},
			{Value: "tr", Freq: 547},
			{Value: "ts", Freq: 11},
			{Value: "tw", Freq: 6},
			{Value: "v", Freq: 542},
			{Value: "vl", Freq: 5},
			{Value: "w", Freq: 528},
			{Value: "wh", Freq: 22},
			{Value: "wr", Freq: 11},
			{Value: "x", Freq: 93},
			{Value: "xz", Freq: 24},
			{Value: "z", Freq: 772},
			{Value: "zh", Freq: 22},
		},
		{
			{Value: "b", Freq: 780},
			{Value: "bb", Freq: 65},
			{Value: "bd", Freq: 126},
			{Value: "bh", Freq: 34},
			{Value: "bl", Freq: 23},
			{Value: "bn", Freq: 5},
			{Value: "br", Freq: 183},
			{Value: "bs", Freq: 12},
			{Value: "c", Freq: 964},
			{Value: "cc", Freq: 35},
			{Value: "cch", Freq: 7},
			{Value: "ch", Freq: 421},
			{Value: "chl", Freq: 17},
			{Value: "chm", Freq: 7},
			{Value: "ck", Freq: 168},
			{Value: "ckh", Freq: 6},
			{Value: "ckl", Freq: 29},
			{Value: "cks", Freq: 13},
			{Value: "ckst", Freq: 6},
			{Value: "cl", Freq: 27},
			{Value: "cq", Freq: 34},
			{Value: "cr", Freq: 5},
			{Value: "cs", Freq: 10},
			{Value: "ct", Freq: 39},
			{Value: "d", Freq: 2011},
			{Value: "db", Freq: 8},
			{Value: "dd", Freq: 154},
			{Value: "ddh", Freq: 7},
			{Value: "ddr", Freq: 13},
			{Value: "df", Freq: 13},
			{Value: "dg", Freq: 61},
			{Value: "dh", Freq: 49},
			{Value: "dj", Freq: 9},
			{Value: "dl", Freq: 39},
			{Value: "dm", Freq: 33},
			{Value: "dn", Freq: 28},
			{Value: "dr", Freq: 266},
			{Value: "ds", Freq: 19},
			{Value: "dv", Freq: 13},
			{Value: "dw", Freq: 55},
			{Value: "f", Freq: 252},
			{Value: "ff", Freq: 80},
			{Value: "ffr", Freq: 24},
			{Value: "fr", Freq: 35},
			{Value: "ft", Freq: 23},
			{Value: "g", Freq: 538},
			{Value: "gd", Freq: 13},
			{Value: "gfr", Freq: 5},
			{Value: "gg", Freq: 32},
			{Value: "gh", Freq: 33},
			{Value: "ght", Freq: 16},
			{Value: "gl", Freq: 12},
			{Value: "gm", Freq: 16},
			{Value: "gn", Freq: 30},
			{Value: "gr", Freq: 23},
			{Value: "h", Freq: 706},
			{Value: "hc", Freq: 6},
			{Value: "hd", Freq: 18},
			{Value: "hj", Freq: 19},
			{Value: "hk", Freq: 33},
			{Value: "hl", Freq: 57},
			{Value: "hm", Freq: 128},
			{Value: "hn", Freq: 63},
			{Value: "hnm", Freq: 5},
			{Value: "hnn", Freq: 13},
			{Value: "hnr", Freq: 6},
			{Value: "hnt", Freq: 12},
			{Value: "hq", Freq: 7},
			{Value: "hr", Freq: 57},
			{Value: "hs", Freq: 39},
			{Value: "hsh", Freq: 12},
			{Value: "ht", Freq: 11},
			{Value: "hv", Freq: 9},
			{Value: "hw", Freq: 5},
			{Value: "hz", Freq: 19},
			{Value: "j", Freq: 551},
			{Value: "jh", Freq: 12},
			{Value: "jm", Freq: 7},
			{Value: "jv", Freq: 5},
			{Value: "k", Freq: 1191},
			{Value: "kh", Freq: 79},
			{Value: "kk", Freq: 37},
			{Value: "kl", Freq: 25},
			{Value: "kr", Freq: 18},
			{Value: "ks", Freq: 30},
			{Value: "ksh", Freq: 13},
			{Value: "kt", Freq: 8},
			{Value: "kw", Freq: 60},
			{Value: "l", Freq: 3533},
			{Value: "lb", Freq: 160},
			{Value: "lc", Freq: 34},
			{Value: "lch", Freq: 7},
			{Value: "ld", Freq: 199},
			{Value: "ldr", Freq: 33},
			{Value: "lf", Freq: 63},
			{Value: "lfr", Freq: 23},
			{Value: "lg", Freq: 35},
			{Value: "lh", Freq: 32},
			{Value: "lj", Freq: 24},
			{Value: "lk", Freq: 26},
			{Value: "ll", Freq: 738},
			{Value: "llm", Freq: 10},
			{Value: "lm", Freq: 172},
			{Value: "ln", Freq: 12},
			{Value: "lp", Freq: 8},
			{Value: "lph", Freq: 40},
			{Value: "lq", Freq: 9},
			{Value: "lr", Freq: 56},
			{Value: "ls", Freq: 59},
			{Value: "lsh", Freq: 9},
			{Value: "lst", Freq: 16},
			{Value: "lt", Freq: 114},
			{Value: "lth", Freq: 6},
			{Value: "lv", Freq: 182},
			{Value: "lw", Freq: 24},
			{Value: "lz", Freq: 10},
			{Value: "m", Freq: 3136},
			{Value: "mb", Freq: 56},
			{Value: "mbr", Freq: 18},
			{Value: "md", Freq: 17},
			{Value: "mh", Freq: 5},
			{Value: "mj", Freq: 15},
			{Value: "ml", Freq: 5},
			{Value: "mm", Freq: 168},
			{Value: "mn", Freq: 12},
			{Value: "mp", Freq: 11},
			{Value: "mps", Freq: 6},
			{Value: "mpt", Freq: 6},
			{Value: "mr", Freq: 77},
			{Value: "ms", Freq: 29},
			{Value: "mt", Freq: 5},
			{Value: "mv", Freq: 6},
			{Value: "mz", Freq: 16},
			{Value: "n", Freq: 3183},
			{Value: "nb", Freq: 17},
			{Value: "nc", Freq: 312},
			{Value: "nch", Freq: 22},
			{Value: "nchr", Freq: 7},
			{Value: "ncl", Freq: 5},
			{Value: "nd", Freq: 682},
			{Value: "ndl", Freq: 25},
			{Value: "ndr", Freq: 332},
			{Value: "nds", Freq: 7},
			{Value: "nf", Freq: 39},
			{Value: "nfr", Freq: 10},
			{Value: "ng", Freq: 131},
			{Value: "ngl", Freq: 14},
			{Value: "ngm", Freq: 6},
			{Value: "ngst", Freq: 11},
			{Value: "ngt", Freq: 23},
			{Value: "nh", Freq: 28},
			{Value: "nj", Freq: 130},
			{Value: "nk", Freq: 59},
			{Value: "nkl", Freq: 6},
			{Value: "nl", Freq: 84},
			{Value: "nm", Freq: 67},
			{Value: "nn", Freq: 641},
			{Value: "nnd", Freq: 9},
			{Value: "nnl", Freq: 9},
			{Value: "np", Freq: 20},
			{Value: "npr", Freq: 5},
			{Value: "nq", Freq: 21},
			{Value: "nr", Freq: 74},
			{Value: "nrr", Freq: 5},
			{Value: "ns", Freq: 153},
			{Value: "nsf", Freq: 5},
			{Value: "nsh", Freq: 19},
			{Value: "nsl", Freq: 18},
			{Value: "nst", Freq: 32},
			{Value: "nt", Freq: 944},
			{Value: "nth", Freq: 98},
			{Value: "ntl", Freq: 25},
			{Value: "ntr", Freq: 137},
			{Value: "ntw", Freq: 34},
			{Value: "nv", Freq: 38},
			{Value: "nw", Freq: 24},
			{Value: "nz", Freq: 151},
			{Value: "p", Freq: 158},
			{Value: "ph", Freq: 154},
			{Value: "phr", Freq: 8},
			{Value: "pl", Freq: 13},
			{Value: "pp", Freq: 37},
			{Value: "pr", Freq: 20},
			{Value: "pt", Freq: 16},
			{Value: "q", Freq: 425},
			{Value: "qw", Freq: 20},
			{Value: "r", Freq: 4648},
			{Value: "rb", Freq: 95},
			{Value: "rc", Freq: 164},
			{Value: "rch", Freq: 34},
			{Value: "rcq", Freq: 5},
			{Value: "rd", Freq: 335},
			{Value: "rf", Freq: 16},
			{Value: "rg", Freq: 115},
			{Value: "rh", Freq: 34},
			{Value: "rj", Freq: 42},
			{Value: "rk", Freq: 146},
			{Value: "rkl", Freq: 12},
			{Value: "rl", Freq: 394},
			{Value: "rlt", Freq: 5},
			{Value: "rm", Freq: 306},
			{Value: "rn", Freq: 306},
			{Value: "rnc", Freq: 5},
			{Value: "rp", Freq: 12},
			{Value: "rph", Freq: 8},
			{Value: "rq", Freq: 97},
			{Value: "rr", Freq: 903},
			{Value: "rs", Freq: 118},
			{Value: "rsh", Freq: 58},
			{Value: "rsk", Freq: 5},
			{Value: "rst", Freq: 22},
			{Value: "rt", Freq: 297},
			{Value: "rth", Freq: 48},
			{Value: "rtl", Freq: 27},
			{Value: "rtn", Freq: 9},
			{Value: "rtr", Freq: 17},
			{Value: "rv", Freq: 232},
			{Value: "rw", Freq: 57},
			{Value: "rz", Freq: 23},
			{Value: "s", Freq: 1609},
			{Value: "sb", Freq: 31},
			{Value: "sc", Freq: 67},
			{Value: "sch", Freq: 12},
			{Value: "sd", Freq: 15},
			{Value: "sf", Freq: 6},
			{Value: "sg", Freq: 13},
			{Value: "sh", Freq: 801},
			{Value: "shd", Freq: 9},
			{Value: "shl", Freq: 15},
			{Value: "shm", Freq: 18},
			{Value: "shn", Freq: 11},
			{Value: "shr", Freq: 9},
			{Value: "sht", Freq: 24},
			{Value: "shv", Freq: 8},
			{Value: "shw", Freq: 12},
			{Value: "sj", Freq: 15},
			{Value: "sk", Freq: 36},
			{Value: "sl", Freq: 87},
			{Value: "sm", Freq: 89},
			{Value: "sn", Freq: 20},
			{Value: "sp", Freq: 40},
			{Value: "sq", Freq: 8},
			{Value: "sr", Freq: 18},
			{Value: "ss", Freq: 306},
			{Value: "ssl", Freq: 6},
			{Value: "sst", Freq: 7},
			{Value: "st", Freq: 744},
			{Value: "sth", Freq: 14},
			{Value: "stl", Freq: 12},
			{Value: "str", Freq: 16},
			{Value: "sv", Freq: 12},
			{Value: "sw", Freq: 15},
			{Value: "sz", Freq: 8},
			{Value: "t", Freq: 898},
			{Value: "tch", Freq: 24},
			{Value: "th", Freq: 389},
			{Value: "thr", Freq: 9},
			{Value: "thv", Freq: 5},
			{Value: "tl", Freq: 35},
			{Value: "tm", Freq: 10},
			{Value: "tn", Freq: 8},
			{Value: "tr", Freq: 213},
			{Value: "ts", Freq: 44},
			{Value: "tt", Freq: 176},
			{Value: "tth", Freq: 43},
			{Value: "ttl", Freq: 13},
			{Value: "tv", Freq: 7},
			{Value: "tw", Freq: 16},
			{Value: "tz", Freq: 21},
			{Value: "v", Freq: 2396},
			{Value: "vl", Freq: 14},
			{Value: "vn", Freq: 6},
			{Value: "vr", Freq: 53},
			{Value: "vv", Freq: 8},
			{Value: "w", Freq: 406},
			{Value: "wd", Freq: 12},
			{Value: "wf", Freq: 5},
			{Value: "wh", Freq: 5},
			{Value: "wj", Freq: 7},
			{Value: "wk", Freq: 8},
			{Value: "wl", Freq: 24},
			{Value: "wn", Freq: 13},
			{Value: "wnt", Freq: 8},
			{Value: "wr", Freq: 13},
			{Value: "ws", Freq: 10},
			{Value: "x", Freq: 202},
			{Value: "xc", Freq: 6},
			{Value: "xd", Freq: 5},
			{Value: "xl", Freq: 12},
			{Value: "xs", Freq: 18},
			{Value: "xst", Freq: 10},
			{Value: "xt", Freq: 70},
			{Value: "xx", Freq: 15},
			{Value: "xz", Freq: 14},
			{Value: "z", Freq: 615},
			{Value: "zd", Freq: 9},
			{Value: "zh", Freq: 16},
			{Value: "zj", Freq: 5},
			{Value: "zk", Freq: 7},
			{Value: "zl", Freq: 6},
			{Value: "zm", Freq: 33},
			{Value: "zr", Freq: 23},
			{Value: "zt", Freq: 7},
			{Value: "zz", Freq: 34},
		},
		{
			{Value: "b", Freq: 232},
			{Value: "bb", Freq: 11},
			{Value: "bh", Freq: 6},
			{Value: "c", Freq: 227},
			{Value: "cc", Freq: 5},
			{Value: "ch", Freq: 61},
			{Value: "ck", Freq: 356},
			{Value: "cks", Freq: 7},
			{Value: "d", Freq: 705},
			{Value: "dd", Freq: 25},
			{Value: "dh", Freq: 9},
			{Value: "dn", Freq: 10},
			{Value: "f", Freq: 123},
			{Value: "ff", Freq: 26},
			{Value: "g", Freq: 64},
			{Value: "gg", Freq: 12},
			{Value: "gh", Freq: 38},
			{Value: "ghn", Freq: 33},
			{Value: "ght", Freq: 10},
			{Value: "h", Freq: 591},
			{Value: "hd", Freq: 5},
			{Value: "hl", Freq: 13},
			{Value: "hm", Freq: 8},
			{Value: "hn", Freq: 85},
			{Value: "hs", Freq: 5},
			{Value: "j", Freq: 89},
			{Value: "jr", Freq: 7},
			{Value: "k", Freq: 450},
			{Value: "kh", Freq: 8},
			{Value: "ks", Freq: 9},
			{Value: "ksh", Freq: 9},
			{Value: "l", Freq: 2297},
			{Value: "ld", Freq: 129},
			{Value: "lf", Freq: 12},
			{Value: "ll", Freq: 737},
			{Value: "lm", Freq: 8},
			{Value: "ln", Freq: 8},
			{Value: "lph", Freq: 11},
			{Value: "ls", Freq: 14},
			{Value: "lt", Freq: 15},
			{Value: "lz", Freq: 5},
			{Value: "m", Freq: 853},
			{Value: "mm", Freq: 11},
			{Value: "mp", Freq: 6},
			{Value: "ms", Freq: 11},
			{Value: "n", Freq: 10007},
			{Value: "nc", Freq: 8},
			{Value: "nch", Freq: 7},
			{Value: "nd", Freq: 268},
			{Value: "ndr", Freq: 5},
			{Value: "ndt", Freq: 5},
			{Value: "ng", Freq: 223},
			{Value: "nh", Freq: 19},
			{Value: "nk", Freq: 22},
			{Value: "nn", Freq: 259},
			{Value: "ns", Freq: 55},
			{Value: "nsh", Freq: 36},
			{Value: "nt", Freq: 140},
			{Value: "nth", Freq: 24},
			{Value: "ntz", Freq: 7},
			{Value: "nz", Freq: 13},
			{Value: "p", Freq: 93},
			{Value: "ph", Freq: 52},
			{Value: "pp", Freq: 9},
			{Value: "q", Freq: 57},
			{Value: "r", Freq: 2423},
			{Value: "rc", Freq: 5},
			{Value: "rch", Freq: 6},
			{Value: "rd", Freq: 358},
			{Value: "rdt", Freq: 5},
			{Value: "rg", Freq: 8},
			{Value: "rk", Freq: 32},
			{Value: "rl", Freq: 57},
			{Value: "rld", Freq: 6},
			{Value: "rn", Freq: 98},
			{Value: "rr", Freq: 38},
			{Value: "rs", Freq: 27},
			{Value: "rsh", Freq: 17},
			{Value: "rt", Freq: 160},
			{Value: "rth", Freq: 34},
			{Value: "rv", Freq: 6},
			{Value: "s", Freq: 2593},
			{Value: "sh", Freq: 214},
			{Value: "ss", Freq: 108},
			{Value: "st", Freq: 41},
			{Value: "sz", Freq: 11},
			{Value: "t", Freq: 358},
			{Value: "tch", Freq: 10},
			{Value: "th", Freq: 176},
			{Value: "ts", Freq: 5},
			{Value: "tt", Freq: 132},
			{Value: "tz", Freq: 9},
			{Value: "v", Freq: 121},
			{Value: "vn", Freq: 7},
			{Value: "w", Freq: 111},
			{Value: "wn", Freq: 144},
			{Value: "ws", Freq: 5},
			{Value: "x", Freq: 122},
			{Value: "xx", Freq: 14},
			{Value: "z", Freq: 282},
			{Value: "zz", Freq: 7},
		},
	},
}).Generator()
