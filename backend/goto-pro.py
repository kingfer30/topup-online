import requests

headers = {
    'accept': '*/*',
    'accept-language': 'zh-CN,zh;q=0.9,en-US;q=0.8,en;q=0.7,en-GB;q=0.6',
    'cache-control': 'no-cache',
    'pragma': 'no-cache',
    'referer': 'https://cursor.com/',
    'sec-ch-ua': '"Not:A-Brand";v="99", "Microsoft Edge";v="145", "Chromium";v="145"',
    'sec-ch-ua-mobile': '?0',
    'sec-ch-ua-platform': '"Windows"',
    'sec-fetch-dest': 'script',
    'sec-fetch-mode': 'no-cors',
    'sec-fetch-site': 'cross-site',
    'sec-fetch-storage-access': 'none',
    'user-agent': 'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/145.0.0.0 Safari/537.36 Edg/145.0.0.0',
}

response = requests.get(
    'https://js.stripe.com/basil/stripe.js', headers=headers)

print(response.text)

headers = {
    'accept': 'text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.7',
    'accept-language': 'zh-CN,zh;q=0.9,en-US;q=0.8,en;q=0.7,en-GB;q=0.6',
    'cache-control': 'no-cache',
    'pragma': 'no-cache',
    'priority': 'u=0, i',
    'referer': 'https://cursor.com/',
    'sec-ch-ua': '"Not:A-Brand";v="99", "Microsoft Edge";v="145", "Chromium";v="145"',
    'sec-ch-ua-mobile': '?0',
    'sec-ch-ua-platform': '"Windows"',
    'sec-fetch-dest': 'iframe',
    'sec-fetch-mode': 'navigate',
    'sec-fetch-site': 'cross-site',
    'sec-fetch-storage-access': 'none',
    'sec-fetch-user': '?1',
    'upgrade-insecure-requests': '1',
    'user-agent': 'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/145.0.0.0 Safari/537.36 Edg/145.0.0.0',
}

response = requests.get(
    'https://js.stripe.com/v3/controller-with-preconnect-562954400a1d5ed269a33dd401040ace.html',
    headers=headers,
)
print(response.text)

headers = {
    'accept': 'text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.7',
    'accept-language': 'zh-CN,zh;q=0.9,en-US;q=0.8,en;q=0.7,en-GB;q=0.6',
    'cache-control': 'no-cache',
    'pragma': 'no-cache',
    'priority': 'u=0, i',
    'referer': 'https://cursor.com/',
    'sec-ch-ua': '"Not:A-Brand";v="99", "Microsoft Edge";v="145", "Chromium";v="145"',
    'sec-ch-ua-mobile': '?0',
    'sec-ch-ua-platform': '"Windows"',
    'sec-fetch-dest': 'iframe',
    'sec-fetch-mode': 'navigate',
    'sec-fetch-site': 'cross-site',
    'sec-fetch-storage-access': 'none',
    'sec-fetch-user': '?1',
    'upgrade-insecure-requests': '1',
    'user-agent': 'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/145.0.0.0 Safari/537.36 Edg/145.0.0.0',
}

response = requests.get(
    'https://js.stripe.com/v3/m-outer-3437aaddcdf6922d623e172c2d6f9278.html', headers=headers)

print(response.text)

headers = {
    'accept': '*/*',
    'accept-language': 'zh-CN,zh;q=0.9,en-US;q=0.8,en;q=0.7,en-GB;q=0.6',
    'cache-control': 'no-cache',
    'pragma': 'no-cache',
    'referer': 'https://js.stripe.com/v3/controller-with-preconnect-562954400a1d5ed269a33dd401040ace.html',
    'sec-ch-ua': '"Not:A-Brand";v="99", "Microsoft Edge";v="145", "Chromium";v="145"',
    'sec-ch-ua-mobile': '?0',
    'sec-ch-ua-platform': '"Windows"',
    'sec-fetch-dest': 'script',
    'sec-fetch-mode': 'no-cors',
    'sec-fetch-site': 'same-origin',
    'sec-fetch-storage-access': 'none',
    'user-agent': 'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/145.0.0.0 Safari/537.36 Edg/145.0.0.0',
}

response = requests.get(
    'https://js.stripe.com/v3/fingerprinted/js/shared-9a246c2de00bc656633679d45127b572.js',
    headers=headers,
)

print(response.text)

headers = {
    'accept': '*/*',
    'accept-language': 'zh-CN,zh;q=0.9,en-US;q=0.8,en;q=0.7,en-GB;q=0.6',
    'cache-control': 'no-cache',
    'pragma': 'no-cache',
    'referer': 'https://js.stripe.com/v3/controller-with-preconnect-562954400a1d5ed269a33dd401040ace.html',
    'sec-ch-ua': '"Not:A-Brand";v="99", "Microsoft Edge";v="145", "Chromium";v="145"',
    'sec-ch-ua-mobile': '?0',
    'sec-ch-ua-platform': '"Windows"',
    'sec-fetch-dest': 'script',
    'sec-fetch-mode': 'no-cors',
    'sec-fetch-site': 'same-origin',
    'sec-fetch-storage-access': 'none',
    'user-agent': 'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/145.0.0.0 Safari/537.36 Edg/145.0.0.0',
}

response = requests.get(
    'https://js.stripe.com/v3/fingerprinted/js/controller-with-preconnect-01e2ee01eaf8224dfc7590fb8fc67129.js',
    headers=headers,
)

print(response.text)

headers = {
    'accept': '*/*',
    'accept-language': 'zh-CN,zh;q=0.9,en-US;q=0.8,en;q=0.7,en-GB;q=0.6',
    'cache-control': 'no-cache',
    'pragma': 'no-cache',
    'referer': 'https://js.stripe.com/v3/m-outer-3437aaddcdf6922d623e172c2d6f9278.html',
    'sec-ch-ua': '"Not:A-Brand";v="99", "Microsoft Edge";v="145", "Chromium";v="145"',
    'sec-ch-ua-mobile': '?0',
    'sec-ch-ua-platform': '"Windows"',
    'sec-fetch-dest': 'script',
    'sec-fetch-mode': 'no-cors',
    'sec-fetch-site': 'same-origin',
    'sec-fetch-storage-access': 'none',
    'user-agent': 'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/145.0.0.0 Safari/537.36 Edg/145.0.0.0',
}

response = requests.get(
    'https://js.stripe.com/v3/fingerprinted/js/m-outer-15a2b40a058ddff1cffdb63779fe3de1.js',
    headers=headers,
)

print(response.text)

cookies = {
    '__stripe_mid': '3c9ed7e3-dbe0-4f3d-b077-cc6ccc2795385258f7',
    'cursor_anonymous_id': 'f6f71ebc-a301-48f6-b4ea-444742000807',
    'statsig_stable_id': '10e8acfe-8bed-40a6-907e-3040018aa0d8',
    '_ca_device_id': 'ca_25d49fda-a256-4a1f-afa0-4f88ad179149',
    'generaltranslation.locale-routing-enabled': 'true',
    'generaltranslation.referrer-locale': 'en-US',
    'workos_id': 'user_01KJ5GCK9BTGKZKWYPGR0245WT',
    'WorkosCursorSessionToken': 'user_01KJ5GCK9BTGKZKWYPGR0245WT::eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiJhdXRoMHx1c2VyXzAxS0o1R0NLOUJUR0taS1dZUEdSMDI0NVdUIiwidGltZSI6IjE3NzE4NTkwMTciLCJyYW5kb21uZXNzIjoiMzhlZjIyOTItYmE0YS00YzQ0IiwiZXhwIjoxNzc3MDQzMDE3LCJpc3MiOiJodHRwczovL2F1dGhlbnRpY2F0aW9uLmN1cnNvci5zaCIsInNjb3BlIjoib3BlbmlkIHByb2ZpbGUgZW1haWwgb2ZmbGluZV9hY2Nlc3MiLCJhdWQiOiJodHRwczovL2N1cnNvci5jb20iLCJ0eXBlIjoic2Vzc2lvbiJ9.Qcg2qR54WyV2CMQDZGQlYJQ-T4BFKFziXnHYgFlvtn8',
    '__stripe_sid': 'f3a94b2a-6ecf-4273-9d03-dc80950c2a7d323cdc',
}

headers = {
    'accept': '*/*',
    'accept-language': 'zh-CN,zh;q=0.9,en-US;q=0.8,en;q=0.7,en-GB;q=0.6',
    'cache-control': 'no-cache',
    'content-type': 'application/json',
    'origin': 'https://cursor.com',
    'pragma': 'no-cache',
    'priority': 'u=1, i',
    'referer': 'https://cursor.com/dashboard?tab=billing',
    'sec-ch-ua': '"Not:A-Brand";v="99", "Microsoft Edge";v="145", "Chromium";v="145"',
    'sec-ch-ua-arch': '"x86"',
    'sec-ch-ua-bitness': '"64"',
    'sec-ch-ua-mobile': '?0',
    'sec-ch-ua-platform': '"Windows"',
    'sec-ch-ua-platform-version': '"19.0.0"',
    'sec-fetch-dest': 'empty',
    'sec-fetch-mode': 'cors',
    'sec-fetch-site': 'same-origin',
    'user-agent': 'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/145.0.0.0 Safari/537.36 Edg/145.0.0.0',
    # 'cookie': '__stripe_mid=3c9ed7e3-dbe0-4f3d-b077-cc6ccc2795385258f7; cursor_anonymous_id=f6f71ebc-a301-48f6-b4ea-444742000807; statsig_stable_id=10e8acfe-8bed-40a6-907e-3040018aa0d8; _ca_device_id=ca_25d49fda-a256-4a1f-afa0-4f88ad179149; generaltranslation.locale-routing-enabled=true; generaltranslation.referrer-locale=en-US; workos_id=user_01KHDEVKTENYK12ZHTEQJ4D6X6; WorkosCursorSessionToken=user_01KHDEVKTENYK12ZHTEQJ4D6X6%3A%3AeyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiJhdXRoMHx1c2VyXzAxS0hERVZLVEVOWUsxMlpIVEVRSjRENlg2IiwidGltZSI6IjE3NzI1OTUyOTciLCJyYW5kb21uZXNzIjoiZjBiOTZkZjgtYzgwYy00MDVkIiwiZXhwIjoxNzc3Nzc5Mjk3LCJpc3MiOiJodHRwczovL2F1dGhlbnRpY2F0aW9uLmN1cnNvci5zaCIsInNjb3BlIjoib3BlbmlkIHByb2ZpbGUgZW1haWwgb2ZmbGluZV9hY2Nlc3MiLCJhdWQiOiJodHRwczovL2N1cnNvci5jb20iLCJ0eXBlIjoid2ViIiwid29ya29zU2Vzc2lvbklkIjoic2Vzc2lvbl8wMUtKVkVLRUFDV0hXQUpKMjRZN0JWTlozWSJ9.t9r5x8CM2VgNeMQevS46J41liHetEtqiBVvvl3bbnvo; __stripe_sid=f3a94b2a-6ecf-4273-9d03-dc80950c2a7d323cdc',
}

json_data = {
    'tier': 'pro',
    'allowAutomaticPayment': True,
    'yearly': False,
}

response = requests.post('https://cursor.com/api/checkout',
                         cookies=cookies, headers=headers, json=json_data)

print(response.text)
