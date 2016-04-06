常规使用方式

- 新建 config 和 template 文件夹

这是模板可以添加gzip头并检查`Content-Encoding`头
template格式如下:

```yml
- hash: gzip
  request:
    header:
      Accept-Encoding: gzip, deflate, sdch
  requirement:
    header:
      Content-Encoding:
      - method: match
        obj: gzip
      - method: show if exist
```

其中hash的test为id, 可以被config中调用

这个配置包括了gzip检查并添加了`Cache-Control`检查
config格式如下:


```yml
- hash: https://github.com
  request:
    hostname: github.com
    uri: /
    method: GET
    scheme: https
    useragent: Mozilla/5.0 (Macintosh; Intel Mac OS X 10_10_5) AppleWebKit/537.36
      (KHTML, like Gecko) Chrome/49.0.2623.87 Safari/537.36
    keepalive: false
    skiptls: true
    compression: false
    timeout: 30s
    include:
    - gzip
    header:
      Cache-Control: max-age=0
  requirement:
    statuscode: 200
    include:
    - gzip
    header:
      Cache-Control:
      - obj: no-cache
        method: include
        option:
        - ignore case
```

```text
./ghtest -c example_config -t example_template -ip 192.30.252.130
192.30.252.130 example_config/github.com.yml:https://github.com https://github.com/ 344.676483ms
- [✓] resp code: 200
- [✓] Content-Encoding match gzip 
- [✓] Content-Encoding show if exist  gzip
- [✓] Cache-Control include no-cache

./ghtest -c example_config -t example_template
github.com example_config/github.com.yml:https://github.com https://github.com/ 1.10781954s
- [✓] resp code: 200
- [✓] Content-Encoding match gzip 
- [✓] Content-Encoding show if exist  gzip
- [✓] Cache-Control include no-cache

./ghtest -c example_config -t example_template -raw -curl
github.com example_config/github.com.yml:https://github.com https://github.com/
200 OK
Cache-Control: no-cache
Content-Encoding: gzip
Content-Security-Policy: default-src 'none'; base-uri 'self'; block-all-mixed-content; child-src render.githubusercontent.com; connect-src 'self' uploads.github.com status.github.com api.github.com www.google-analytics.com github-cloud.s3.amazonaws.com api.braintreegateway.com client-analytics.braintreegateway.com wss://live.github.com; font-src assets-cdn.github.com; form-action 'self' github.com gist.github.com; frame-ancestors 'none'; frame-src render.githubusercontent.com; img-src 'self' data: assets-cdn.github.com identicons.github.com www.google-analytics.com collector.githubapp.com *.gravatar.com *.wp.com checkout.paypal.com *.githubusercontent.com; media-src 'none'; object-src assets-cdn.github.com; plugin-types application/x-shockwave-flash; script-src assets-cdn.github.com; style-src 'unsafe-inline' assets-cdn.github.com
Content-Type: text/html; charset=utf-8
Date: Wed, 06 Apr 2016 22:21:11 GMT
Public-Key-Pins: max-age=5184000; pin-sha256="WoiWRyIOVNa9ihaBciRSC7XHjliYS9VwUGOIud4PB18="; pin-sha256="RRM1dGqnDFsCJXBTHky16vi1obOlCgFFn/yOhI/y+ho="; pin-sha256="k2v657xBsOVe1PQRwOsHsw3bsGT2VzIqz5K+59sNQws="; pin-sha256="K87oWBWM9UZfyddvDfoxL+8lpNyoUB2ptGtn0fv6G2Q="; pin-sha256="IQBnNBEiFuhj+8x6X8XLgh01V9Ic5/V3IRQLNFFc7v4="; pin-sha256="iie1VXtL7HzAMF+/PVPR9xzT80kQxdZeJ+zduCB3uj0="; pin-sha256="LvRiGEjRqfzurezaWuj8Wie2gyHMrW5Q06LspMnox7A="; includeSubDomains
Server: GitHub.com
Set-Cookie: logged_in=no; domain=.github.com; path=/; expires=Sun, 06 Apr 2036 22:21:11 -0000; secure; HttpOnly
Status: 200 OK
Strict-Transport-Security: max-age=31536000; includeSubdomains; preload
Vary: X-PJAX
X-Content-Type-Options: nosniff
X-Frame-Options: deny
X-Github-Request-Id: 182B0C2A:2081:236FE7C:57058BD6
X-Request-Id: bb51559fdfed1f8256c0768fbff34639
X-Runtime: 0.013500
X-Served-By: 926b734ea1992f8ee1f88ab967a93dac
X-Ua-Compatible: IE=Edge,chrome=1
X-Xss-Protection: 1; mode=block

github.com example_config/github.com.yml:https://github.com https://github.com/ 3.014893527s
curl -s -k -X GET -H Host:github.com -A 'Mozilla/5.0 (Macintosh; Intel Mac OS X 10_10_5) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/49.0.2623.87 Safari/537.36' -H 'Accept-Encoding: gzip, deflate, sdch' -H 'Cache-Control: max-age=0' 'https://github.com/'
- [✓] resp code: 200
- [✓] Content-Encoding match gzip 
- [✓] Content-Encoding show if exist  gzip
- [✓] Cache-Control include no-cache
```

- ip 可以使用`,`分隔, 比如`ghtest -ip 1.1.1.1,2.2.2.2`
- 在需要检查curl模拟环境是可以使用`-curl`
- 在需要得到原始返回头文件时可以使用`-raw`