# Утилиты

Генерация сертификатов

```
make crt
```

# Результат профилирования через pprof

## Условия
Профилирование проводилось для следующиъ эндпоинтов

```
POST /                  - генерация короткого URL 1
POST /api/shorten       - генерация короткого URL 2
POST /api/shorten/batch - генерация коротких URL
GET  /{id}              - получение короткого URL (с драйвером файл)
```

Были сняты snapshot до рефакторинга и после. Сначала была создана нагрузка через hey

```
hey -m POST -d "https://abc1.loc/" -z 5s http://127.0.0.1:8080
hey -m POST -H "Accept: application/json" -d "{\"url\": \"http://bla.loc\"}" -z 5s http://127.0.0.1:8080/api/shorten 
hey -m POST -H "Accept: application/json" -d "[{\"correlation_id\": \"abc1\",\"original_url\": \"http://abb1.loc\"},{\"correlation_id\": \"abc2\",\"original_url\": \"http://abb2.loc\"}]
" -z 5s http://127.0.0.1:8080/api/shorten/batch
hey -m POST -d "https://google.com/" -z 1s http://127.0.0.1:8080
hey -z 5s http://127.0.0.1:8080/I4wLZred
```

Затем снят snapshot

```
curl -sK -v http://localhost:6060/debug/pprof/heap > base.pprof
```

И уже после выполнил профилирование

```
go tool pprof -http=":9090" -seconds=30 profiles/base.pprof
```

Затем сделал рефакторинг, и повторно выполнил условтя. После второг оснапшота сравнил их
```
go tool pprof -http=:9090 -diff_base=profiles/base.pprof profiles/result.pprof
```

Проверку проводил на предмет alloc_space

## Проблемные места

### Получение короткого URL

Поскольку профилирование проводилось с использованием сервисного драйвера File, удалось обнаружить неоптимальные решения.

У основной структуры записи данных из файла был изменен тип одног ополя

```
-       IsDeleted string `json:"is_deleted"`
+       IsDeleted bool   `json:"is_deleted"`
```

И убран везде парсинг булева значения

```
-               b, err := strconv.ParseBool(item.IsDeleted)
-               if err != nil {
-                       return nil, err
-               }
```

Также местами была заменена мапа на структуру

```
-               rec := map[string]string{}
+               rec := recordFile{}
```

При сравнении профилей удалось заметить значительное улучшение в контексте alloc_space. storage.GetByShort стал прогонять через себя значительно меньше памяти -20.50MB

### генерация короткого URL 1

Для формирования полной ссылки использовался fmt.Sprintf

```
full := fmt.Sprintf("%s/%s", h.conf.BaseUrl, id)
```

Планировщик в режиме alloc_space показал не лучший резулитат. 

После изменения на

```
full := h.conf.BaseUrl + "/" + id
```

Есть небольшие улучшения: diff -1.50MB

### Маппер MapGenUrlsResp

В мапере MapGenUrlsResp вместо

```
res := []models.ShortenBatchResp{}
```

Я ограничил емкость через

```
res := make([]models.ShortenBatchResp, 0, len(urls))
```

Как я понял, этот подход через append не должен вызывать realloc

Также в маппере я заменил формирование ссылки с

```
full := fmt.Sprintf("%s/%s", baseUrl, url.Short)
```

на

```
full := baseUrl + "/" + url.Short
```

Проведя diff заметил небольшое улучшение в -1MB

## Результат

### Команда

```
go tool pprof -sample_index=alloc_space -top -diff_base=profiles/base.pprof profiles/result.pprof
```

### Результат с фильтром

```
...
-20.50MB  0.04%  0.58%      -31MB  0.06%  github.com/spider4216/tinyurl/internal/storage.(*FileStorage).GetByShort             
-1.50MB 0.0029%  0.61%  1737.13MB  3.37%  github.com/spider4216/tinyurl/internal/handler.Handler.GenerateId         
-1MB 0.0019%  0.61%       -2MB 0.0039%  github.com/spider4216/tinyurl/internal/service.Service.MapForMapUrlIds        
-1MB 0.0019%  0.61%    -5.50MB 0.011%  github.com/spider4216/tinyurl/internal/handler.Handler.MapGenUrlsResp
-1MB 0.0019%  0.61% -2812.41MB  5.46%  github.com/spider4216/tinyurl/internal/handler.Handler.GetShortenUrls
...
```

# Бенчмарки

Бенчмарки были написаны на два основных метода генерации коротких URL

```
POST /                  - генерация короткого URL
POST /api/shorten/batch - генерация коротких URL
```