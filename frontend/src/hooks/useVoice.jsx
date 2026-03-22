import { useCallback, useEffect } from 'react';

let audioContext = null;
let source = null;

// 初始化 AudioContext
const initAudioContext = () => {
    if (!audioContext) {
        audioContext = new (window.AudioContext || window.webkitAudioContext)();
    }

    // 某些浏览器需要用户交互后才能启动 AudioContext
    if (audioContext.state === 'suspended') {
        audioContext.resume().then(() => {
        });
    }

    return audioContext;
};

export const useVoice = () => {
    useEffect(() => {
        // 组件挂载时初始化
        initAudioContext();
    }, []);

    const playOrStop = useCallback(async (data, options = {}) => {
        try {
            const ctx = initAudioContext();

            if (source) {
                // 如果正在播放，停止播放
                source.stop();
                source.disconnect();
                source = null;
                return;
            }

            // 如果没在播放，开始播放

            // 确保 AudioContext 是运行状态
            if (ctx.state === 'suspended') {
                await ctx.resume();
            }

            // 使用 Promise 版本的 decodeAudioData
            const buffer = await ctx.decodeAudioData(data.buffer);

            source = ctx.createBufferSource();
            source.buffer = buffer;

            // 设置播放速度
            if (options.rate) {
                source.playbackRate.value = options.rate;
            } else {
                // 默认速度稍微调快一点，通常 1.1 - 1.2 倍速比较自然
                source.playbackRate.value = 1.2;
            }

            source.connect(ctx.destination);

            source.start(0);

            source.onended = () => {
                if (source) {
                    source.disconnect();
                    source = null;
                }
            };

            // 返回 Promise 等待播放完成
            return new Promise((resolve) => {
                const onEnd = () => {
                    resolve();
                };
                if (source) {
                    source.addEventListener('ended', onEnd, { once: true });
                } else {
                    resolve();
                }
            });
        } catch (err) {
            console.error('音频播放错误:', err);
            console.error('错误详情:', err.message, err.stack);
            if (source) {
                try {
                    source.disconnect();
                } catch (e) {
                    // 忽略断开连接错误
                }
                source = null;
            }
            throw err; // 抛出错误让调用者处理
        }
    }, []);

    return playOrStop;
};
