// Copyright ©2017-2020 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package namer

import "github.com/richardwilkes/rpgtools/names"

// Female is a random name generator for female first names. Created from data
// obtained from http://www.ssa.gov/oact/babynames/names.zip
var Female = (&names.Data{
	StartsWithVowelFreq:     12795,
	StartsWithConsonantFreq: 52863,
	CountFreq:               []int{0, 1002, 3580, 18158, 15383, 21032, 4147, 1986, 187, 108, 44, 14, 17},
	Segments: [names.ArraySize][]names.Segment{
		{
			{Value: "a", Freq: 6306},
			{Value: "aa", Freq: 217},
			{Value: "aai", Freq: 9},
			{Value: "aay", Freq: 5},
			{Value: "aaya", Freq: 6},
			{Value: "ae", Freq: 59},
			{Value: "ai", Freq: 220},
			{Value: "aiya", Freq: 8},
			{Value: "ao", Freq: 8},
			{Value: "au", Freq: 293},
			{Value: "ay", Freq: 142},
			{Value: "aya", Freq: 32},
			{Value: "aye", Freq: 15},
			{Value: "ayo", Freq: 13},
			{Value: "ayu", Freq: 5},
			{Value: "e", Freq: 2162},
			{Value: "ea", Freq: 75},
			{Value: "ee", Freq: 8},
			{Value: "ei", Freq: 66},
			{Value: "eu", Freq: 94},
			{Value: "ey", Freq: 18},
			{Value: "i", Freq: 951},
			{Value: "ia", Freq: 17},
			{Value: "ie", Freq: 5},
			{Value: "io", Freq: 16},
			{Value: "iy", Freq: 12},
			{Value: "iya", Freq: 14},
			{Value: "iyo", Freq: 7},
			{Value: "o", Freq: 678},
			{Value: "oa", Freq: 14},
			{Value: "ou", Freq: 6},
			{Value: "u", Freq: 162},
			{Value: "y", Freq: 69},
			{Value: "ya", Freq: 468},
			{Value: "yaa", Freq: 5},
			{Value: "yae", Freq: 9},
			{Value: "yai", Freq: 29},
			{Value: "ye", Freq: 124},
			{Value: "yei", Freq: 15},
			{Value: "yi", Freq: 41},
			{Value: "yo", Freq: 115},
			{Value: "yoa", Freq: 6},
			{Value: "you", Freq: 7},
			{Value: "yu", Freq: 137},
		},
		{
			{Value: "a", Freq: 30416},
			{Value: "aa", Freq: 253},
			{Value: "aai", Freq: 7},
			{Value: "ae", Freq: 654},
			{Value: "aea", Freq: 11},
			{Value: "aei", Freq: 7},
			{Value: "aeo", Freq: 9},
			{Value: "aey", Freq: 6},
			{Value: "aeya", Freq: 7},
			{Value: "ai", Freq: 1683},
			{Value: "aia", Freq: 86},
			{Value: "aie", Freq: 17},
			{Value: "aio", Freq: 15},
			{Value: "aiy", Freq: 22},
			{Value: "aiya", Freq: 62},
			{Value: "ao", Freq: 56},
			{Value: "aoi", Freq: 5},
			{Value: "au", Freq: 623},
			{Value: "ay", Freq: 1761},
			{Value: "aya", Freq: 298},
			{Value: "aye", Freq: 105},
			{Value: "ayi", Freq: 17},
			{Value: "ayia", Freq: 21},
			{Value: "ayo", Freq: 55},
			{Value: "ayu", Freq: 8},
			{Value: "ayya", Freq: 6},
			{Value: "e", Freq: 21239},
			{Value: "ea", Freq: 1508},
			{Value: "eaa", Freq: 7},
			{Value: "eai", Freq: 28},
			{Value: "eau", Freq: 43},
			{Value: "eay", Freq: 11},
			{Value: "ee", Freq: 1404},
			{Value: "eea", Freq: 64},
			{Value: "eeo", Freq: 9},
			{Value: "eeya", Freq: 18},
			{Value: "ei", Freq: 1334},
			{Value: "eia", Freq: 63},
			{Value: "eio", Freq: 23},
			{Value: "eiya", Freq: 13},
			{Value: "eo", Freq: 391},
			{Value: "eu", Freq: 32},
			{Value: "ey", Freq: 409},
			{Value: "eya", Freq: 196},
			{Value: "eye", Freq: 20},
			{Value: "eyi", Freq: 7},
			{Value: "eyo", Freq: 57},
			{Value: "eyu", Freq: 6},
			{Value: "i", Freq: 16506},
			{Value: "ia", Freq: 2387},
			{Value: "iaa", Freq: 17},
			{Value: "iae", Freq: 8},
			{Value: "iai", Freq: 18},
			{Value: "iau", Freq: 35},
			{Value: "iay", Freq: 6},
			{Value: "iaya", Freq: 17},
			{Value: "ie", Freq: 1316},
			{Value: "iea", Freq: 103},
			{Value: "ieo", Freq: 12},
			{Value: "ieya", Freq: 5},
			{Value: "ii", Freq: 23},
			{Value: "io", Freq: 462},
			{Value: "iou", Freq: 18},
			{Value: "iu", Freq: 50},
			{Value: "iy", Freq: 87},
			{Value: "iya", Freq: 647},
			{Value: "iye", Freq: 14},
			{Value: "iyi", Freq: 5},
			{Value: "iyia", Freq: 7},
			{Value: "iyo", Freq: 50},
			{Value: "iyya", Freq: 20},
			{Value: "o", Freq: 8164},
			{Value: "oa", Freq: 83},
			{Value: "oe", Freq: 139},
			{Value: "oea", Freq: 11},
			{Value: "oei", Freq: 5},
			{Value: "oey", Freq: 5},
			{Value: "oi", Freq: 91},
			{Value: "oo", Freq: 125},
			{Value: "ou", Freq: 230},
			{Value: "oua", Freq: 9},
			{Value: "oue", Freq: 10},
			{Value: "oui", Freq: 24},
			{Value: "oy", Freq: 94},
			{Value: "oya", Freq: 27},
			{Value: "oye", Freq: 11},
			{Value: "oyi", Freq: 7},
			{Value: "oyo", Freq: 5},
			{Value: "u", Freq: 2594},
			{Value: "ua", Freq: 350},
			{Value: "uai", Freq: 7},
			{Value: "uay", Freq: 5},
			{Value: "ue", Freq: 325},
			{Value: "uea", Freq: 9},
			{Value: "uee", Freq: 24},
			{Value: "ui", Freq: 320},
			{Value: "uia", Freq: 8},
			{Value: "uie", Freq: 11},
			{Value: "uo", Freq: 30},
			{Value: "uy", Freq: 18},
			{Value: "uye", Freq: 10},
			{Value: "y", Freq: 6194},
			{Value: "ya", Freq: 669},
			{Value: "yai", Freq: 31},
			{Value: "yau", Freq: 15},
			{Value: "ye", Freq: 225},
			{Value: "yea", Freq: 11},
			{Value: "yee", Freq: 9},
			{Value: "yei", Freq: 11},
			{Value: "yi", Freq: 38},
			{Value: "yia", Freq: 88},
			{Value: "yie", Freq: 11},
			{Value: "yio", Freq: 8},
			{Value: "yo", Freq: 137},
			{Value: "yu", Freq: 18},
			{Value: "yya", Freq: 5},
		},
		{
			{Value: "a", Freq: 19678},
			{Value: "aa", Freq: 58},
			{Value: "aaya", Freq: 9},
			{Value: "ae", Freq: 544},
			{Value: "aea", Freq: 13},
			{Value: "aee", Freq: 5},
			{Value: "aeya", Freq: 11},
			{Value: "ai", Freq: 195},
			{Value: "aia", Freq: 62},
			{Value: "aie", Freq: 7},
			{Value: "aii", Freq: 7},
			{Value: "aiya", Freq: 80},
			{Value: "ao", Freq: 13},
			{Value: "ay", Freq: 311},
			{Value: "aya", Freq: 340},
			{Value: "aye", Freq: 57},
			{Value: "ayi", Freq: 6},
			{Value: "ayia", Freq: 33},
			{Value: "ayo", Freq: 25},
			{Value: "ayya", Freq: 6},
			{Value: "e", Freq: 7168},
			{Value: "ea", Freq: 640},
			{Value: "eaya", Freq: 5},
			{Value: "ee", Freq: 1419},
			{Value: "eea", Freq: 18},
			{Value: "eeya", Freq: 24},
			{Value: "ei", Freq: 154},
			{Value: "eia", Freq: 97},
			{Value: "eiya", Freq: 19},
			{Value: "eo", Freq: 13},
			{Value: "ey", Freq: 1005},
			{Value: "eya", Freq: 146},
			{Value: "eyia", Freq: 14},
			{Value: "i", Freq: 3306},
			{Value: "ia", Freq: 4443},
			{Value: "iaa", Freq: 7},
			{Value: "iaya", Freq: 26},
			{Value: "ie", Freq: 2450},
			{Value: "iea", Freq: 27},
			{Value: "iee", Freq: 36},
			{Value: "ieya", Freq: 14},
			{Value: "ii", Freq: 84},
			{Value: "iia", Freq: 6},
			{Value: "io", Freq: 35},
			{Value: "iy", Freq: 6},
			{Value: "iya", Freq: 440},
			{Value: "iye", Freq: 18},
			{Value: "iyia", Freq: 8},
			{Value: "iyo", Freq: 10},
			{Value: "iyya", Freq: 10},
			{Value: "o", Freq: 469},
			{Value: "oa", Freq: 34},
			{Value: "oe", Freq: 39},
			{Value: "oee", Freq: 6},
			{Value: "oey", Freq: 12},
			{Value: "oi", Freq: 21},
			{Value: "oia", Freq: 5},
			{Value: "oie", Freq: 11},
			{Value: "oo", Freq: 8},
			{Value: "ou", Freq: 70},
			{Value: "oua", Freq: 15},
			{Value: "oy", Freq: 35},
			{Value: "oya", Freq: 61},
			{Value: "oye", Freq: 16},
			{Value: "oyia", Freq: 9},
			{Value: "u", Freq: 110},
			{Value: "ua", Freq: 135},
			{Value: "uay", Freq: 5},
			{Value: "ue", Freq: 211},
			{Value: "uea", Freq: 6},
			{Value: "ui", Freq: 12},
			{Value: "uia", Freq: 17},
			{Value: "uie", Freq: 6},
			{Value: "uoia", Freq: 7},
			{Value: "uoya", Freq: 5},
			{Value: "uy", Freq: 5},
			{Value: "uye", Freq: 8},
			{Value: "uyo", Freq: 10},
			{Value: "y", Freq: 2415},
			{Value: "ya", Freq: 639},
			{Value: "yae", Freq: 8},
			{Value: "ye", Freq: 190},
			{Value: "yea", Freq: 8},
			{Value: "yi", Freq: 9},
			{Value: "yia", Freq: 112},
			{Value: "yya", Freq: 5},
		},
		{
			{Value: "b", Freq: 1074},
			{Value: "bh", Freq: 14},
			{Value: "bl", Freq: 97},
			{Value: "br", Freq: 1040},
			{Value: "c", Freq: 2222},
			{Value: "ch", Freq: 1248},
			{Value: "chl", Freq: 20},
			{Value: "chr", Freq: 183},
			{Value: "cl", Freq: 294},
			{Value: "cn", Freq: 5},
			{Value: "cr", Freq: 150},
			{Value: "d", Freq: 3720},
			{Value: "dh", Freq: 18},
			{Value: "dhr", Freq: 5},
			{Value: "dj", Freq: 15},
			{Value: "dl", Freq: 11},
			{Value: "dm", Freq: 9},
			{Value: "dn", Freq: 20},
			{Value: "dr", Freq: 68},
			{Value: "dw", Freq: 15},
			{Value: "dz", Freq: 9},
			{Value: "f", Freq: 549},
			{Value: "fl", Freq: 114},
			{Value: "fr", Freq: 168},
			{Value: "g", Freq: 973},
			{Value: "gh", Freq: 17},
			{Value: "gl", Freq: 144},
			{Value: "gr", Freq: 175},
			{Value: "gw", Freq: 79},
			{Value: "h", Freq: 1116},
			{Value: "hr", Freq: 6},
			{Value: "j", Freq: 4479},
			{Value: "jh", Freq: 62},
			{Value: "jk", Freq: 5},
			{Value: "jl", Freq: 14},
			{Value: "jm", Freq: 9},
			{Value: "jn", Freq: 15},
			{Value: "k", Freq: 4287},
			{Value: "kh", Freq: 239},
			{Value: "khl", Freq: 11},
			{Value: "khr", Freq: 24},
			{Value: "kl", Freq: 48},
			{Value: "km", Freq: 8},
			{Value: "kn", Freq: 15},
			{Value: "kr", Freq: 239},
			{Value: "kw", Freq: 16},
			{Value: "l", Freq: 4407},
			{Value: "ll", Freq: 27},
			{Value: "m", Freq: 5356},
			{Value: "mcc", Freq: 6},
			{Value: "mck", Freq: 103},
			{Value: "mcl", Freq: 5},
			{Value: "mh", Freq: 5},
			{Value: "mk", Freq: 9},
			{Value: "ml", Freq: 6},
			{Value: "mr", Freq: 6},
			{Value: "n", Freq: 2417},
			{Value: "nd", Freq: 6},
			{Value: "ng", Freq: 13},
			{Value: "nh", Freq: 10},
			{Value: "nk", Freq: 14},
			{Value: "p", Freq: 760},
			{Value: "ph", Freq: 125},
			{Value: "pl", Freq: 12},
			{Value: "pr", Freq: 173},
			{Value: "q", Freq: 268},
			{Value: "qw", Freq: 6},
			{Value: "r", Freq: 2640},
			{Value: "rh", Freq: 134},
			{Value: "s", Freq: 2695},
			{Value: "sc", Freq: 22},
			{Value: "sch", Freq: 26},
			{Value: "sh", Freq: 2748},
			{Value: "shl", Freq: 5},
			{Value: "shn", Freq: 7},
			{Value: "shr", Freq: 35},
			{Value: "shw", Freq: 5},
			{Value: "sk", Freq: 85},
			{Value: "sl", Freq: 6},
			{Value: "sm", Freq: 11},
			{Value: "sn", Freq: 12},
			{Value: "sp", Freq: 15},
			{Value: "sr", Freq: 44},
			{Value: "st", Freq: 230},
			{Value: "str", Freq: 6},
			{Value: "sv", Freq: 6},
			{Value: "sw", Freq: 26},
			{Value: "t", Freq: 3502},
			{Value: "th", Freq: 211},
			{Value: "thr", Freq: 11},
			{Value: "tk", Freq: 12},
			{Value: "tm", Freq: 5},
			{Value: "tn", Freq: 7},
			{Value: "tr", Freq: 439},
			{Value: "ts", Freq: 18},
			{Value: "tw", Freq: 29},
			{Value: "tz", Freq: 14},
			{Value: "v", Freq: 1068},
			{Value: "vr", Freq: 5},
			{Value: "w", Freq: 515},
			{Value: "wh", Freq: 27},
			{Value: "wr", Freq: 15},
			{Value: "x", Freq: 140},
			{Value: "z", Freq: 1060},
			{Value: "zh", Freq: 35},
			{Value: "zn", Freq: 5},
		},
		{
			{Value: "b", Freq: 1166},
			{Value: "bb", Freq: 114},
			{Value: "bbr", Freq: 8},
			{Value: "bh", Freq: 20},
			{Value: "bl", Freq: 24},
			{Value: "bn", Freq: 9},
			{Value: "br", Freq: 359},
			{Value: "bs", Freq: 5},
			{Value: "c", Freq: 2407},
			{Value: "cc", Freq: 63},
			{Value: "ch", Freq: 505},
			{Value: "chl", Freq: 6},
			{Value: "chr", Freq: 9},
			{Value: "ck", Freq: 268},
			{Value: "ckl", Freq: 23},
			{Value: "ckq", Freq: 6},
			{Value: "cl", Freq: 68},
			{Value: "cq", Freq: 78},
			{Value: "cr", Freq: 38},
			{Value: "cs", Freq: 12},
			{Value: "ct", Freq: 52},
			{Value: "d", Freq: 3191},
			{Value: "dd", Freq: 176},
			{Value: "ddh", Freq: 6},
			{Value: "ddl", Freq: 8},
			{Value: "ddr", Freq: 5},
			{Value: "dg", Freq: 30},
			{Value: "dh", Freq: 53},
			{Value: "dj", Freq: 15},
			{Value: "dl", Freq: 62},
			{Value: "dm", Freq: 18},
			{Value: "dn", Freq: 68},
			{Value: "dr", Freq: 348},
			{Value: "ds", Freq: 9},
			{Value: "dv", Freq: 8},
			{Value: "dw", Freq: 29},
			{Value: "f", Freq: 325},
			{Value: "ff", Freq: 115},
			{Value: "ffn", Freq: 8},
			{Value: "ffr", Freq: 5},
			{Value: "fh", Freq: 5},
			{Value: "fn", Freq: 11},
			{Value: "fr", Freq: 38},
			{Value: "fs", Freq: 7},
			{Value: "ft", Freq: 15},
			{Value: "g", Freq: 583},
			{Value: "gd", Freq: 23},
			{Value: "gg", Freq: 37},
			{Value: "gh", Freq: 100},
			{Value: "ghl", Freq: 20},
			{Value: "ghn", Freq: 5},
			{Value: "ght", Freq: 14},
			{Value: "gl", Freq: 14},
			{Value: "gn", Freq: 59},
			{Value: "gr", Freq: 49},
			{Value: "h", Freq: 834},
			{Value: "hb", Freq: 7},
			{Value: "hd", Freq: 12},
			{Value: "hj", Freq: 35},
			{Value: "hk", Freq: 13},
			{Value: "hl", Freq: 163},
			{Value: "hm", Freq: 74},
			{Value: "hn", Freq: 232},
			{Value: "hnn", Freq: 32},
			{Value: "hr", Freq: 77},
			{Value: "hs", Freq: 7},
			{Value: "ht", Freq: 10},
			{Value: "hv", Freq: 9},
			{Value: "hz", Freq: 19},
			{Value: "j", Freq: 1027},
			{Value: "jh", Freq: 34},
			{Value: "jl", Freq: 7},
			{Value: "jm", Freq: 5},
			{Value: "jn", Freq: 5},
			{Value: "jsh", Freq: 5},
			{Value: "k", Freq: 3048},
			{Value: "kh", Freq: 83},
			{Value: "kk", Freq: 85},
			{Value: "kl", Freq: 51},
			{Value: "kq", Freq: 7},
			{Value: "kr", Freq: 21},
			{Value: "ks", Freq: 17},
			{Value: "ksh", Freq: 34},
			{Value: "kt", Freq: 16},
			{Value: "kth", Freq: 5},
			{Value: "kw", Freq: 15},
			{Value: "l", Freq: 12802},
			{Value: "lb", Freq: 107},
			{Value: "lbr", Freq: 9},
			{Value: "lc", Freq: 60},
			{Value: "ld", Freq: 271},
			{Value: "ldr", Freq: 19},
			{Value: "lf", Freq: 24},
			{Value: "lfr", Freq: 19},
			{Value: "lg", Freq: 25},
			{Value: "lh", Freq: 27},
			{Value: "lj", Freq: 15},
			{Value: "lk", Freq: 34},
			{Value: "ll", Freq: 2960},
			{Value: "lls", Freq: 14},
			{Value: "lm", Freq: 178},
			{Value: "ln", Freq: 34},
			{Value: "lp", Freq: 9},
			{Value: "lph", Freq: 34},
			{Value: "lr", Freq: 36},
			{Value: "ls", Freq: 114},
			{Value: "lsh", Freq: 8},
			{Value: "lss", Freq: 5},
			{Value: "lt", Freq: 66},
			{Value: "lth", Freq: 15},
			{Value: "lv", Freq: 185},
			{Value: "lw", Freq: 22},
			{Value: "lz", Freq: 34},
			{Value: "m", Freq: 4485},
			{Value: "mb", Freq: 168},
			{Value: "mbl", Freq: 9},
			{Value: "mbr", Freq: 92},
			{Value: "md", Freq: 14},
			{Value: "mh", Freq: 5},
			{Value: "mk", Freq: 8},
			{Value: "ml", Freq: 26},
			{Value: "mm", Freq: 291},
			{Value: "mn", Freq: 29},
			{Value: "mp", Freq: 32},
			{Value: "mph", Freq: 10},
			{Value: "mpl", Freq: 6},
			{Value: "mpr", Freq: 9},
			{Value: "mr", Freq: 101},
			{Value: "ms", Freq: 29},
			{Value: "mt", Freq: 6},
			{Value: "mz", Freq: 10},
			{Value: "n", Freq: 13799},
			{Value: "nb", Freq: 18},
			{Value: "nc", Freq: 390},
			{Value: "nch", Freq: 43},
			{Value: "nd", Freq: 1611},
			{Value: "ndh", Freq: 13},
			{Value: "ndl", Freq: 31},
			{Value: "ndr", Freq: 568},
			{Value: "nds", Freq: 24},
			{Value: "ndz", Freq: 11},
			{Value: "nf", Freq: 14},
			{Value: "ng", Freq: 283},
			{Value: "ngl", Freq: 21},
			{Value: "ngr", Freq: 16},
			{Value: "ngst", Freq: 5},
			{Value: "ngt", Freq: 23},
			{Value: "nh", Freq: 23},
			{Value: "nj", Freq: 132},
			{Value: "nk", Freq: 76},
			{Value: "nkl", Freq: 5},
			{Value: "nl", Freq: 212},
			{Value: "nm", Freq: 36},
			{Value: "nn", Freq: 3648},
			{Value: "nnd", Freq: 16},
			{Value: "nndr", Freq: 6},
			{Value: "nnl", Freq: 46},
			{Value: "nnm", Freq: 6},
			{Value: "nnz", Freq: 5},
			{Value: "nq", Freq: 32},
			{Value: "nr", Freq: 50},
			{Value: "ns", Freq: 160},
			{Value: "nsh", Freq: 42},
			{Value: "nsl", Freq: 65},
			{Value: "nst", Freq: 15},
			{Value: "nt", Freq: 744},
			{Value: "nth", Freq: 102},
			{Value: "ntl", Freq: 27},
			{Value: "ntr", Freq: 72},
			{Value: "nts", Freq: 5},
			{Value: "ntw", Freq: 14},
			{Value: "nv", Freq: 53},
			{Value: "nw", Freq: 19},
			{Value: "nz", Freq: 217},
			{Value: "nzl", Freq: 44},
			{Value: "p", Freq: 204},
			{Value: "ph", Freq: 201},
			{Value: "phn", Freq: 10},
			{Value: "phr", Freq: 8},
			{Value: "pl", Freq: 11},
			{Value: "pp", Freq: 40},
			{Value: "pph", Freq: 5},
			{Value: "pr", Freq: 66},
			{Value: "ps", Freq: 15},
			{Value: "pt", Freq: 11},
			{Value: "q", Freq: 715},
			{Value: "qw", Freq: 6},
			{Value: "r", Freq: 9755},
			{Value: "rb", Freq: 93},
			{Value: "rbr", Freq: 8},
			{Value: "rc", Freq: 163},
			{Value: "rch", Freq: 43},
			{Value: "rd", Freq: 337},
			{Value: "rdr", Freq: 7},
			{Value: "rf", Freq: 10},
			{Value: "rg", Freq: 235},
			{Value: "rgh", Freq: 7},
			{Value: "rgr", Freq: 16},
			{Value: "rh", Freq: 45},
			{Value: "rj", Freq: 49},
			{Value: "rk", Freq: 122},
			{Value: "rkl", Freq: 16},
			{Value: "rl", Freq: 1151},
			{Value: "rld", Freq: 13},
			{Value: "rll", Freq: 7},
			{Value: "rm", Freq: 404},
			{Value: "rn", Freq: 575},
			{Value: "rp", Freq: 15},
			{Value: "rph", Freq: 8},
			{Value: "rq", Freq: 61},
			{Value: "rr", Freq: 1204},
			{Value: "rrl", Freq: 6},
			{Value: "rs", Freq: 161},
			{Value: "rsch", Freq: 8},
			{Value: "rsh", Freq: 102},
			{Value: "rsl", Freq: 5},
			{Value: "rst", Freq: 86},
			{Value: "rt", Freq: 272},
			{Value: "rth", Freq: 113},
			{Value: "rtl", Freq: 34},
			{Value: "rtn", Freq: 40},
			{Value: "rtr", Freq: 24},
			{Value: "rv", Freq: 142},
			{Value: "rw", Freq: 18},
			{Value: "rz", Freq: 41},
			{Value: "s", Freq: 3422},
			{Value: "sb", Freq: 29},
			{Value: "sc", Freq: 76},
			{Value: "sch", Freq: 34},
			{Value: "sd", Freq: 11},
			{Value: "sf", Freq: 5},
			{Value: "sh", Freq: 3194},
			{Value: "shb", Freq: 5},
			{Value: "shk", Freq: 21},
			{Value: "shl", Freq: 78},
			{Value: "shm", Freq: 23},
			{Value: "shn", Freq: 18},
			{Value: "shr", Freq: 12},
			{Value: "sht", Freq: 23},
			{Value: "shv", Freq: 10},
			{Value: "shw", Freq: 8},
			{Value: "sj", Freq: 19},
			{Value: "sk", Freq: 47},
			{Value: "sl", Freq: 266},
			{Value: "sm", Freq: 184},
			{Value: "sn", Freq: 28},
			{Value: "sp", Freq: 26},
			{Value: "sq", Freq: 5},
			{Value: "sr", Freq: 40},
			{Value: "ss", Freq: 1048},
			{Value: "ssh", Freq: 12},
			{Value: "ssl", Freq: 30},
			{Value: "ssm", Freq: 11},
			{Value: "st", Freq: 966},
			{Value: "sth", Freq: 16},
			{Value: "stl", Freq: 19},
			{Value: "stn", Freq: 5},
			{Value: "str", Freq: 25},
			{Value: "sv", Freq: 11},
			{Value: "sw", Freq: 10},
			{Value: "sz", Freq: 8},
			{Value: "t", Freq: 2901},
			{Value: "tc", Freq: 5},
			{Value: "tch", Freq: 14},
			{Value: "th", Freq: 840},
			{Value: "thl", Freq: 47},
			{Value: "thm", Freq: 7},
			{Value: "thr", Freq: 38},
			{Value: "thz", Freq: 18},
			{Value: "tl", Freq: 129},
			{Value: "tm", Freq: 7},
			{Value: "tn", Freq: 46},
			{Value: "tr", Freq: 437},
			{Value: "ts", Freq: 88},
			{Value: "tsh", Freq: 6},
			{Value: "tt", Freq: 1340},
			{Value: "ttl", Freq: 16},
			{Value: "ttn", Freq: 28},
			{Value: "tv", Freq: 5},
			{Value: "tz", Freq: 177},
			{Value: "v", Freq: 2572},
			{Value: "vd", Freq: 5},
			{Value: "vl", Freq: 17},
			{Value: "vn", Freq: 10},
			{Value: "vr", Freq: 49},
			{Value: "vv", Freq: 9},
			{Value: "w", Freq: 443},
			{Value: "wd", Freq: 9},
			{Value: "wl", Freq: 6},
			{Value: "wn", Freq: 132},
			{Value: "wnd", Freq: 22},
			{Value: "wndr", Freq: 9},
			{Value: "wnn", Freq: 12},
			{Value: "wnt", Freq: 29},
			{Value: "wr", Freq: 13},
			{Value: "ws", Freq: 10},
			{Value: "x", Freq: 327},
			{Value: "xc", Freq: 11},
			{Value: "xl", Freq: 24},
			{Value: "xs", Freq: 22},
			{Value: "xt", Freq: 13},
			{Value: "xx", Freq: 17},
			{Value: "xz", Freq: 11},
			{Value: "z", Freq: 1106},
			{Value: "zb", Freq: 12},
			{Value: "zh", Freq: 44},
			{Value: "zj", Freq: 8},
			{Value: "zl", Freq: 113},
			{Value: "zm", Freq: 82},
			{Value: "zn", Freq: 7},
			{Value: "zr", Freq: 18},
			{Value: "zs", Freq: 5},
			{Value: "zt", Freq: 9},
			{Value: "zz", Freq: 104},
			{Value: "zzl", Freq: 10},
			{Value: "zzm", Freq: 17},
		},
		{
			{Value: "b", Freq: 41},
			{Value: "c", Freq: 45},
			{Value: "ch", Freq: 7},
			{Value: "ck", Freq: 35},
			{Value: "d", Freq: 154},
			{Value: "f", Freq: 17},
			{Value: "g", Freq: 14},
			{Value: "gh", Freq: 314},
			{Value: "ghn", Freq: 9},
			{Value: "ght", Freq: 7},
			{Value: "h", Freq: 3703},
			{Value: "hl", Freq: 5},
			{Value: "hn", Freq: 9},
			{Value: "ht", Freq: 5},
			{Value: "j", Freq: 22},
			{Value: "k", Freq: 55},
			{Value: "l", Freq: 1428},
			{Value: "ld", Freq: 19},
			{Value: "ll", Freq: 625},
			{Value: "ln", Freq: 20},
			{Value: "m", Freq: 240},
			{Value: "mn", Freq: 5},
			{Value: "n", Freq: 5457},
			{Value: "nd", Freq: 108},
			{Value: "ndr", Freq: 6},
			{Value: "ng", Freq: 122},
			{Value: "nh", Freq: 27},
			{Value: "nn", Freq: 1342},
			{Value: "ns", Freq: 10},
			{Value: "nt", Freq: 26},
			{Value: "nth", Freq: 5},
			{Value: "p", Freq: 27},
			{Value: "q", Freq: 12},
			{Value: "r", Freq: 900},
			{Value: "rd", Freq: 35},
			{Value: "rh", Freq: 6},
			{Value: "rk", Freq: 5},
			{Value: "rl", Freq: 37},
			{Value: "rn", Freq: 38},
			{Value: "rr", Freq: 9},
			{Value: "rs", Freq: 7},
			{Value: "rt", Freq: 29},
			{Value: "rth", Freq: 7},
			{Value: "s", Freq: 1046},
			{Value: "sh", Freq: 55},
			{Value: "ss", Freq: 117},
			{Value: "st", Freq: 24},
			{Value: "t", Freq: 426},
			{Value: "th", Freq: 296},
			{Value: "tt", Freq: 146},
			{Value: "v", Freq: 15},
			{Value: "w", Freq: 32},
			{Value: "wn", Freq: 68},
			{Value: "x", Freq: 68},
			{Value: "xx", Freq: 6},
			{Value: "z", Freq: 179},
		},
	},
}).Generator()
