package aggregation

/*
Regular expression for extracting URLs from https://github.com/mvdan/xurls

Copyright (c) 2015, Daniel Martí. All rights reserved.

Redistribution and use in source and binary forms, with or without
modification, are permitted provided that the following conditions are
met:

   * Redistributions of source code must retain the above copyright
notice, this list of conditions and the following disclaimer.
   * Redistributions in binary form must reproduce the above
copyright notice, this list of conditions and the following disclaimer
in the documentation and/or other materials provided with the
distribution.
   * Neither the name of the copyright holder nor the names of its
contributors may be used to endorse or promote products derived from
this software without specific prior written permission.

THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS
"AS IS" AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT
LIMITED TO, THE IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR
A PARTICULAR PURPOSE ARE DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT
OWNER OR CONTRIBUTORS BE LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL,
SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT
LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES; LOSS OF USE,
DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND ON ANY
THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT
(INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE
OF THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.
*/
import (
	"net/url"
	"regexp"
)

const allowedUcsChar = "¡-ᙿᚁ-\u1fff\u200b-‧\u202a-\u202e‰-⁞\u2060-\u2fff、-\ud7ff豈-﷏ﷰ-\uffef𐀀-\U0001fffd𠀀-\U0002fffd𰀀-\U0003fffd\U00040000-\U0004fffd\U00050000-\U0005fffd\U00060000-\U0006fffd\U00070000-\U0007fffd\U00080000-\U0008fffd\U00090000-\U0009fffd\U000a0000-\U000afffd\U000b0000-\U000bfffd\U000c0000-\U000cfffd\U000d0000-\U000dfffd\U000e1000-\U000efffd"
const allowedUcsCharMinusPunctuation = "¢-¦¨-µ¸-¾À-ͽͿ-ΆΈ-ՙՠ-ֈ֊-ֿׁ-ׂׄ-ׇׅ-ײ\u05f5-؈؋؎-ؚ\u061cؠ-٩ٮ-ۓە-ۿ\u070e-߶ߺ-\u082f\u083f-\u085d\u085f-ॣ०-९ॱ-ৼ৾-ੵ\u0a77-૯૱-\u0c76౸-ಃಅ-ෳ\u0df5-๎๐-๙\u0e5c-༃༓༕-྄྆-࿏࿕-࿘\u0fdb-၉ၐ-ჺჼ-፟፩-᙭ᙯ-ᙿᚁ-ᛪᛮ-᜴\u1737-៓ៗ៛-\u17ff᠆᠋-\u1943᥆-\u1a1dᨠ-\u1a9fᪧ\u1aae-᭙᭡-᭼\u1b7f-\u1bfbᰀ-\u1c3a᱀-ᱽᲀ-Ჿ\u1cc8-᳔᳒-\u1fff\u200b-―‘-‟\u202a-\u202e‹-›‿-⁀⁄-⁆⁒⁔\u2060-\u2cf8⳽ⴀ-ⵯ\u2d71-ⷿ⸂-⸅⸉-⸊⸌-⸍⸗⸚⸜-⸝⸠-⸩ⸯ⸺-⸻⹀⹂⹐-⹑⹕-\u2fff〄-〼〾-ヺー-ꓽꔀ-ꘌꘐ-꙲ꙴ-꙽ꙿ-꛱\ua6f8-ꡳ\ua878-\ua8cd꣐-ꣷꣻꣽ-꤭ꤰ-\ua95eꥠ-꧀\ua9ce-\ua9ddꧠ-\uaa5bꩠ-ꫝꫠ-ꫯꫲ-ꯪ꯬-\ud7ff豈-﷏ﷰ-️︗-︘\ufe1a-︯︱-﹄﹇-﹈﹍-﹏\ufe53﹘-﹞﹢-\ufe67﹩\ufe6c-\uff00＄（-）＋－０-９＜-＞Ａ-［］-｠｢-｣ｦ-\uffef𐀀-\U000100ff\U00010103-\U0001039e𐎠-𐏏𐏑-\U0001056e𐕰-\U00010856𐡘-\U0001091e𐤠-\U0001093e\U00010940-\U00010a4f\U00010a59-𐩾𐪀-𐫯\U00010af7-\U00010b38𐭀-\U00010b98\U00010b9d-𐽔\U00010f5a-𐾅\U00010f8a-𑁆\U0001104e-𑂺\U000110bd𑃂-𑄿𑅄-𑅳𑅶-𑇄𑇉-𑇌𑇎-𑇚𑇜\U000111e0-𑈷𑈾-𑊨\U000112aa-𑑊𑑐-𑑙\U0001145c𑑞-𑓅𑓇-𑗀𑗘-𑙀𑙄-\U0001165f\U0001166d-𑚸\U000116ba-𑜻𑜿-𑠺\U0001183c-𑥃\U00011947-𑧡𑧣-𑨾𑩇-𑪙𑪝\U00011aa3-\U00011aff\U00011b0a-𑱀\U00011c46-\U00011c6f𑱲-𑻶\U00011ef9-𑽂𑽐-\U00011ffe𒀀-\U0001246f\U00012475-𒿰\U00012ff3-\U00016a6d𖩰-𖫴\U00016af6-𖬶𖬼-𖭃𖭅-𖺖\U00016e9b-𖿡𖿣-𛲞\U0001bca0-𝪆\U0001da8c-\U0001e95d\U0001e960-\U0001fffd𠀀-\U0002fffd𰀀-\U0003fffd\U00040000-\U0004fffd\U00050000-\U0005fffd\U00060000-\U0006fffd\U00070000-\U0007fffd\U00080000-\U0008fffd\U00090000-\U0009fffd\U000a0000-\U000afffd\U000b0000-\U000bfffd\U000c0000-\U000cfffd\U000d0000-\U000dfffd\U000e1000-\U000efffd"

const (
	unreservedChar      = `a-zA-Z0-9\-._~`
	endUnreservedChar   = `a-zA-Z0-9\-_~`
	midSubDelimChar     = `!$&'*+,;=`
	endSubDelimChar     = `$&+=`
	midIPathSegmentChar = unreservedChar + `%` + midSubDelimChar + `:@` + allowedUcsChar
	endIPathSegmentChar = endUnreservedChar + `%` + endSubDelimChar + allowedUcsCharMinusPunctuation
	iPrivateChar        = `\x{E000}-\x{F8FF}\x{F0000}-\x{FFFFD}\x{100000}-\x{10FFFD}`
	midIChar            = `/?#\\` + midIPathSegmentChar + iPrivateChar
	endIChar            = `/#` + endIPathSegmentChar + iPrivateChar
	wellParen           = `\((?:[` + midIChar + `]|\([` + midIChar + `]*\))*\)`
	wellBracket         = `\[(?:[` + midIChar + `]|\[[` + midIChar + `]*\])*\]`
	wellBrace           = `\{(?:[` + midIChar + `]|\{[` + midIChar + `]*\})*\}`
	wellAll             = wellParen + `|` + wellBracket + `|` + wellBrace
	pathCont            = `(?:[` + midIChar + `]*(?:` + wellAll + `|[` + endIChar + `]))+`
	schemes             = `(?:(?i)(?:http|https)://)`
)

func extractUrls(text string) []string {
	re := regexp.MustCompile(schemes + pathCont)
	re.Longest()

	var urls []string
	for _, match := range re.FindAllString(text, -1) {
		if _, err := url.Parse(match); err == nil {
			urls = append(urls, match)
		}
	}
	return urls
}
