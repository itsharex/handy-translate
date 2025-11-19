export async function tts(text, lang, options = {}) {
    const { config } = options;

    let lingvaConfig = { requestPath: 'lingva.pot-app.com' };

    if (config !== undefined) {
        lingvaConfig = config;
    }

    let { requestPath } = lingvaConfig;
    if (!requestPath.startsWith('http')) {
        requestPath = 'https://' + requestPath;
    }

    console.log('TTS API 请求:', `${requestPath}/api/v1/audio/${lang}/${encodeURIComponent(text)}`);
    const response = await fetch(`${requestPath}/api/v1/audio/${lang}/${encodeURIComponent(text)}`);

    const jsonData = await response.json();
    console.log('TTS API 响应类型:', typeof jsonData);
    // console.log('TTS API 响应:', jsonData); // 避免打印过长的日志

    if (response.ok) {
        let audioData = jsonData['audio'];
        
        if (audioData) {
            // 如果 audioData 是字符串类型 (Base64)
            if (typeof audioData === 'string') {
                // 移除可能的 data URI 前缀
                if (audioData.includes(',')) {
                    const parts = audioData.split(',');
                    audioData = parts[parts.length - 1];
                }
                
                // Base64 解码为 Uint8Array
                try {
                    const binaryString = window.atob(audioData);
                    const len = binaryString.length;
                    const bytes = new Uint8Array(len);
                    for (let i = 0; i < len; i++) {
                        bytes[i] = binaryString.charCodeAt(i);
                    }
                    console.log('Base64 解码成功，数据长度:', bytes.length);
                    return bytes;
                } catch (e) {
                    console.error('Base64 解码失败:', e);
                    throw new Error('TTS 音频数据解码失败');
                }
            }
            // 如果是数组（可能是某些 API 返回了字节数组）
            else if (Array.isArray(audioData)) {
                return new Uint8Array(audioData);
            }
            
            return audioData;
        }
    }

    throw new Error(`TTS API 失败: ${response.status} ${response.statusText}`);
}

export * from './info';
