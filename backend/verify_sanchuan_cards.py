import argparse
import csv
import time
from collections import Counter
from pathlib import Path
from typing import Any

import requests


URI = "https://kkk.ow800.com/api/cards/verify"


def verify_one(card: str, max_retry: int, timeout_sec: int = 20) -> dict[str, Any]:
    for attempt in range(1, max_retry + 1):
        try:
            resp = requests.post(
                URI,
                json={"cardInfo": card},
                timeout=timeout_sec,
            )

            if resp.status_code in (429, 500):
                time.sleep(min(12, 2 * attempt))
                continue

            resp.raise_for_status()
            body = resp.json()
            data = body.get("data") or {}

            return {
                "Card": card,
                "Success": bool(data.get("success", False)),
                "Message": str(data.get("message") or "EMPTY_RESPONSE"),
                "Product": data.get("productName"),
                "Price": data.get("productPrice"),
            }
        except requests.RequestException as exc:
            status = None
            if exc.response is not None:
                status = exc.response.status_code

            if status in (429, 500):
                time.sleep(min(12, 2 * attempt))
                continue

            return {
                "Card": card,
                "Success": False,
                "Message": f"ERROR: {exc}",
                "Product": None,
                "Price": None,
            }
        except ValueError as exc:
            return {
                "Card": card,
                "Success": False,
                "Message": f"ERROR: INVALID_JSON: {exc}",
                "Product": None,
                "Price": None,
            }

    return {
        "Card": card,
        "Success": False,
        "Message": "ERROR: MAX_RETRY_EXCEEDED",
        "Product": None,
        "Price": None,
    }


def main() -> None:
    parser = argparse.ArgumentParser(description="批量验证三川卡密")
    parser.add_argument("--input-file", default="./tmp_cards.txt", help="输入卡密文件")
    parser.add_argument("--valid-file", default="./tmp_valid_cards.txt", help="可用卡密输出文件")
    parser.add_argument("--result-file", default="./tmp_verify_result.csv", help="完整结果输出文件")
    parser.add_argument("--delay-ms", type=int, default=800, help="每次请求间隔毫秒")
    parser.add_argument("--max-retry", type=int, default=5, help="429/500 最大重试次数")
    args = parser.parse_args()

    input_path = Path(args.input_file)
    valid_path = Path(args.valid_file)
    result_path = Path(args.result_file)

    if not input_path.exists():
        raise SystemExit(f"输入文件不存在: {input_path}")

    cards = [line.strip() for line in input_path.read_text(encoding="utf-8").splitlines() if line.strip()]
    if not cards:
        raise SystemExit(f"输入文件为空: {input_path}")

    results: list[dict[str, Any]] = []
    total = len(cards)

    for idx, card in enumerate(cards, start=1):
        result = verify_one(card, max_retry=args.max_retry)
        results.append(result)

        if idx % 10 == 0 or idx == total:
            valid_count = sum(1 for item in results if item["Success"])
            print(f"Progress: {idx}/{total} | Valid: {valid_count}")

        time.sleep(max(0, args.delay_ms) / 1000.0)

    valid = sorted((item for item in results if item["Success"]), key=lambda x: x["Card"])
    invalid = [item for item in results if not item["Success"]]

    valid_path.write_text(
        "\n".join(item["Card"] for item in valid) + ("\n" if valid else ""),
        encoding="utf-8",
    )

    with result_path.open("w", newline="", encoding="utf-8-sig") as f:
        writer = csv.DictWriter(f, fieldnames=["Card", "Success", "Message", "Product", "Price"])
        writer.writeheader()
        writer.writerows(results)

    print("\n===== 验证完成 =====")
    print(f"总数: {total}")
    print(f"可用: {len(valid)}")
    print(f"不可用/异常: {len(invalid)}")
    print(f"可用卡密文件: {valid_path.resolve()}")
    print(f"完整结果文件: {result_path.resolve()}")

    print("\n不可用/异常分类：")
    for message, count in Counter(item["Message"] for item in invalid).most_common():
        print(f"{message}: {count}")


if __name__ == "__main__":
    main()
