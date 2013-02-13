(function() {
  languages = {
    "af": "Afrikaans",
    "sq": "Albanian",
    "ar": "Arabic",
    "az": "Azerbaijani",
    "eu": "Basque",
    "bn": "Bengali",
    "be": "Belarusian",
    "bg": "Bulgarian",
    "ca": "Catalan",
    "zh-CN": "Simplified Chinese",
    "zh-TW": "Traditional Chinese",
    "hr": "Croatian",
    "cs": "Czech",
    "da": "Danish",
    "nl": "Dutch",
    "en": "English",
    "eo": "Esperanto",
    "et": "Estonian",
    "tl": "Filipino",
    "fi": "Finnish",
    "fr": "French",
    "gl": "Galician",
    "ka": "Georgian",
    "de": "German",
    "el": "Greek",
    "gu": "Gujarati",
    "ht": "Haitian Creole",
    "iw": "Hebrew",
    "hi": "Hindi",
    "hu": "Hungarian",
    "is": "Icelandic",
    "id": "Indonesian",
    "ga": "Irish",
    "it": "Italian",
    "ja": "Japanese",
    "kn": "Kannada",
    "ko": "Korean",
    "la": "Latin",
    "lv": "Latvian",
    "lt": "Lithuanian",
    "mk": "Macedonian",
    "ms": "Malay",
    "mt": "Maltese",
    "no": "Norwegian",
    "fa": "Persian",
    "pl": "Polish",
    "pt": "Portuguese",
    "ro": "Romanian",
    "ru": "Russian",
    "sr": "Serbian",
    "sk": "Slovak",
    "sl": "Slovenian",
    "es": "Spanish",
    "sw": "Swahili",
    "sv": "Swedish",
    "ta": "Tamil",
    "te": "Telugu",
    "th": "Thai",
    "tr": "Turkish",
    "uk": "Ukranian",
    "ur": "Urdu",
    "vi": "Vietnamese",
    "cy": "Welsh",
    "yi": "Yiddish"
  };

  getCode = function(language, languages) {
    var code, lang;
    for (code in languages) {
      lang = languages[code];
      if (lang.toLowerCase() === language.toLowerCase()) return code;
    }
  };

  module.exports = function(robot) {
    return robot.respond(/(?:translate)(?: me)?(?:(?: from) ([a-z]*))?(?:(?: (?:in)?to) ([a-z]*))? (.*)/i, function(msg) {
      var origin, target, term;
      term = "\"" + msg.match[3] + "\"";
      origin = msg.match[1] !== void 0 ? getCode(msg.match[1], languages) : 'auto';
      target = msg.match[2] !== void 0 ? getCode(msg.match[2], languages) : 'en';
      return msg.http("https://translate.google.com/translate_a/t").query({
        client: 't',
        hl: 'en',
        multires: 1,
        sc: 1,
        sl: origin,
        ssel: 0,
        tl: target,
        tsel: 0,
        uptl: "en",
        text: term
      }).header('User-Agent', 'Mozilla/5.0').get()(function(err, res, body) {
        var data, language, parsed;
        data = body;
        if (data.length > 4 && data[0] === '[') {
          parsed = eval(data);
          language = languages[parsed[2]];
          parsed = parsed[0] && parsed[0][0] && parsed[0][0][0];
          if (parsed) {
            if (msg.match[2] === void 0) {
              return msg.send("" + term + " is " + language + " for " + parsed);
            } else {
              return msg.send("The " + language + " " + term + " translates as " + parsed + " in " + languages[target]);
            }
          }
        }
      });
    });
  };

}).call(this);
