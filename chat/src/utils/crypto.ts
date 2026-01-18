import CryptoJS from 'crypto-js'

/**
 * 对密码进行SHA256加密
 * @param password 原始密码
 * @returns 加密后的密码（十六进制字符串）
 */
export function encryptPassword(password: string): string {
  return CryptoJS.SHA256(password).toString()
}

/**
 * 对字符串进行MD5加密
 * @param str 原始字符串
 * @returns MD5加密后的字符串
 */
export function md5(str: string): string {
  return CryptoJS.MD5(str).toString()
}

/**
 * 使用AES加密存储数据（用于记住我功能）
 * @param data 要加密的数据
 * @returns 加密后的字符串
 */
export function encryptStorage(data: string): string {
  const SECRET_KEY = 'topup-online-remember-key-2024' // 密钥
  return CryptoJS.AES.encrypt(data, SECRET_KEY).toString()
}

/**
 * 解密存储的数据
 * @param encryptedData 加密的数据
 * @returns 解密后的原始字符串，失败返回空字符串
 */
export function decryptStorage(encryptedData: string): string {
  try {
    const SECRET_KEY = 'topup-online-remember-key-2024' // 密钥
    const bytes = CryptoJS.AES.decrypt(encryptedData, SECRET_KEY)
    return bytes.toString(CryptoJS.enc.Utf8)
  } catch (error) {
    console.error('解密失败:', error)
    return ''
  }
}

