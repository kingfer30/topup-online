<script setup lang="ts">
import { ref, onMounted, nextTick, h, Component } from 'vue'

// Language management
const currentLang = ref('zh')
const isLoading = ref(true)

// FAQ state
const faqOpenStates = ref([false, false, false, false])

// Language dictionary
const langDict = {
  en: {
    page_title: "ChatGPT Automated Top-up Platform",
    nav_home: "Home",
    nav_features: "Features", 
    nav_steps: "How it works",
    nav_faq: "FAQ",
    hero_title: "Fast, Secure, Automated ChatGPT Top-up Platform",
    hero_sub: "No login required, Token only, fully automated 24/7 system",
    hero_btn: "Start Recharge",
    hero_card: "Card Purchase",
    features_title: "Features",
    features_auto: "Instant Top-up",
    features_auto_desc: "Recharge in seconds, no waiting",
    features_official: "Official Top-up",
    features_official_desc: "Using official iOS discounted zones",
    features_security: "Data Security",
    features_security_desc: "Full encryption, privacy guaranteed",
    features_oem: "Unattended",
    features_oem_desc: "24/7 Automated Operation",
    steps_title: "How it works",
    step_1_title: "Get Token",
    step_1_desc: "Get your Token from ChatGPT account",
    step_2_title: "Fill Info",
    step_2_desc: "Paste Token & select plan",
    step_3_title: "Submit",
    step_3_desc: "Automated system processing",
    step_4_title: "Instant Arrival",
    step_4_desc: "Top-up arrives instantly, synced",
    faq_title: "FAQ",
    faq_q1: "Is this official top-up?",
    faq_a1: "We use discounted official iOS subscription APIs, 100% genuine.",
    faq_q2: "Will the card code expire?",
    faq_a2: "The card code is only invalid after a successful top-up.",
    faq_q3: "Do I need to log in?",
    faq_a3: "No. Only your Token is needed, no account/password involved.",
    faq_q4: "What if the top-up fails?",
    faq_a4: "Rarely, a failure can occur due to network. If so, check if you have a subscription; if not, just retry.",
    footer_copyright: "© 2025 ChatGPT Top-up Platform. All rights reserved.",
    footer_tech: "Tech Assurance",
    footer_privacy: "Privacy Policy",
  },
  zh: {
    page_title: "ChatGPT 自动化充值平台",
    nav_home: "首页",
    nav_features: "功能亮点",
    nav_steps: "操作流程",
    nav_faq: "常见问题",
    hero_title: "快速、安全、自动化的 ChatGPT 充值平台",
    hero_sub: "无需登录，仅需 Token，全自动 24 小时充值系统",
    hero_btn: "开始充值",
    hero_card: "卡密购买",
    features_title: "功能亮点",
    features_auto: "自动到账",
    features_auto_desc: "充值秒到，无需等待",
    features_official: "正规充值",
    features_official_desc: "采用官方 iOS 低价区",
    features_security: "数据安全",
    features_security_desc: "全程加密，隐私无忧",
    features_oem: "无人值守",
    features_oem_desc: "24/7全天候自动操作",
    steps_title: "操作流程",
    step_1_title: "获取Token",
    step_1_desc: "从 ChatGPT 账户设置获取 Token",
    step_2_title: "填写信息",
    step_2_desc: "粘贴 Token 并选择套餐",
    step_3_title: "提交充值",
    step_3_desc: "系统自动化处理",
    step_4_title: "自动到账",
    step_4_desc: "充值秒到，自动同步",
    faq_title: "常见问题",
    faq_q1: "这是正规充值吗？",
    faq_a1: "我们使用低价区官方订阅接口进行操作，百分百正规充值。",
    faq_q2: "卡密会失效吗？",
    faq_a2: "卡密只有充值成功，才会失效。",
    faq_q3: "是否需要登录账号？",
    faq_a3: "不需要。我们只需要你的 Token，不涉及账号密码。",
    faq_q4: "充值失败怎么办？",
    faq_a4: "极少数情况下充值会失败，如果遇到，请检查你的账号是否有会员，如果没有在点击充值即可，这是由于网络波动造成的。",
    footer_copyright: "© 2025 ChatGPT 自动充值平台 版权所有",
    footer_tech: "技术保障",
    footer_privacy: "隐私条款",
  }
}

// Modal states
const showTechModal = ref(false)
const showPrivacyModal = ref(false)

// Language detection and management
const detectLang = () => {
  const lang = navigator.language || navigator.userLanguage
  if (lang.startsWith('zh')) return 'zh'
  return 'en'
}

const switchLang = (lang: string) => {
  currentLang.value = lang
  localStorage.setItem('lang', lang)
  useHead({
    title: langDict[lang].page_title
  })
}

const toggleFaq = (index: number) => {
  faqOpenStates.value = faqOpenStates.value.map((state, i) => 
    i === index ? !state : false
  )
}

const scrollToTop = () => {
  window.scrollTo({ top: 0, behavior: 'smooth' })
}

const initParticles = () => {
  if (process.client && window.particlesJS) {
    window.particlesJS("particles-js", {
      particles: {
        number: { value: 60, density: { enable: true, value_area: 800 } },
        color: { value: "#ffffff" },
        shape: { type: "circle" },
        opacity: { value: 0.3 },
        size: { value: 3, random: true },
        line_linked: {
          enable: true,
          distance: 150,
          color: "#ffffff",
          opacity: 0.2,
          width: 1
        },
        move: {
          enable: true,
          speed: 2,
          direction: "none",
          random: false,
          straight: false,
          out_mode: "out"
        }
      },
      interactivity: {
        events: {
          onhover: { enable: true, mode: "repulse" },
          onclick: { enable: true, mode: "push" }
        },
        modes: {
          repulse: { distance: 100, duration: 0.4 },
          push: { particles_nb: 4 }
        }
      },
      retina_detect: true
    })
  }
}

// Flag icon components
const USFlagIcon = () => h('span', { style: 'font-size: 16px;' }, '🇺🇸')
const CNFlagIcon = () => h('span', { style: 'font-size: 16px;' }, '🇨🇳')

// Icon render function (simplified)
function renderIcon(icon: Component) {
  return () => h(icon)
}

// Language options for dropdown
const langOptions = [
  {
    label: 'English',
    key: 'en',
    icon: renderIcon(USFlagIcon)
  },
  {
    label: '中文',
    key: 'zh',
    icon: renderIcon(CNFlagIcon)
  }
]

// Lifecycle
onMounted(async () => {
  // Initialize language
  const savedLang = localStorage.getItem('lang') || detectLang()
  currentLang.value = savedLang
  
  // Set page title
  useHead({
    title: langDict[savedLang].page_title
  })
  
  // Load particles.js
  if (process.client) {
    const script = document.createElement('script')
    script.src = 'https://cdn.jsdelivr.net/npm/particles.js@2.0.0/particles.min.js'
    script.onload = () => {
      nextTick(() => {
        initParticles()
      })
    }
    document.head.appendChild(script)
  }
  
  isLoading.value = false
})

// Computed
const t = computed(() => langDict[currentLang.value])
</script>

<template>
  <div>
    <!-- Particle Background -->
    <div id="particles-js" class="fixed inset-0 -z-10"></div>
    
    <!-- Main Content -->
    <div class="bg-gradient-to-br from-[#7e72d6] via-[#7bb6f8] to-[#e0ecff] text-gray-900 min-h-screen animate-[fadeIn_1s_ease-in-out]">
      
      <!-- Header -->
      <header class="bg-white/10 backdrop-blur-sm p-4 sticky top-0 z-40 shadow-xl">
        <div class="max-w-7xl mx-auto flex justify-between items-center flex-col sm:flex-row gap-3 sm:gap-0">
          <div>
            <button 
              @click="scrollToTop" 
              class="nav-link"
            >
              {{ t.nav_home }}
            </button>
          </div>
          <nav class="flex gap-2 items-center justify-center">
            <a href="#features" class="nav-link">
              {{ t.nav_features }}
            </a>
            <a href="#steps" class="nav-link">
              {{ t.nav_steps }}
            </a>
            <a href="#faq" class="nav-link">
              {{ t.nav_faq }}
            </a>
            <n-dropdown 
              :options="langOptions"
              @select="switchLang"
              trigger="click"
            >
              <n-button size="small" class="ml-2 lang-dropdown-btn">
                <span v-if="currentLang === 'zh'">🇨🇳 中文</span>
                <span v-else>🇺🇸 English</span>
              </n-button>
            </n-dropdown>
          </nav>
        </div>
      </header>

      <!-- Hero Section -->
      <section class="py-20 text-center px-4 relative z-10">
        <img 
          alt="ChatGPT Logo" 
          class="w-16 mx-auto mb-5" 
          src="https://upload.wikimedia.org/wikipedia/commons/0/04/ChatGPT_logo.svg"
        />
        <h2 class="text-4xl sm:text-5xl font-bold mb-4 text-white drop-shadow">
          {{ t.hero_title }}
        </h2>
        <p class="text-xl mb-8 text-white drop-shadow">
          {{ t.hero_sub }}
        </p>
        <div class="flex justify-center gap-3 items-center flex-wrap">
          <n-button 
            type="primary" 
            size="large"
            class="bg-white text-indigo-700 font-semibold px-8 py-3 rounded-full shadow-lg hover:shadow-2xl hover:scale-105 transition-all text-lg"
            tag="a"
            href="https://www.ow520.com/"
            target="_blank"
          >
            {{ t.hero_btn }}
          </n-button>
          <n-button 
            secondary
            size="large"
            class="flex items-center gap-1 px-5 py-3 rounded-full border border-indigo-600 text-indigo-700 font-semibold hover:bg-indigo-50 transition-all text-base"
            tag="a"
            href="http://193.200.134.144?cid=2&mid=2"
            target="_blank"
          >
            <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-width="2" d="M8 10h.01M12 14h.01M16 10h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"/>
            </svg>
            {{ t.hero_card }}
          </n-button>
        </div>
      </section>

      <!-- Features Section -->
      <section class="py-20 bg-white text-gray-900 text-center px-6" id="features">
        <h3 class="text-4xl font-bold mb-10">{{ t.features_title }}</h3>
        <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-8 max-w-6xl mx-auto">
          <!-- OEM - 无人值守 -->
          <n-card class="bg-gradient-to-br from-blue-100 via-indigo-100 to-white hover:scale-105 transition-transform">
            <div class="flex flex-col items-center">
              <span class="mb-2">
                <svg width="32" height="32" fill="none" viewBox="0 0 24 24">
                  <!-- 机器人头部 -->
                  <rect x="7" y="6" width="10" height="8" rx="2" fill="#818CF8"/>
                  <!-- 机器人眼睛 -->
                  <circle cx="9.5" cy="9" r="1" fill="#fff"/>
                  <circle cx="14.5" cy="9" r="1" fill="#fff"/>
                  <!-- 机器人嘴巴 -->
                  <rect x="10.5" y="11.5" width="3" height="0.5" rx="0.25" fill="#fff"/>
                  <!-- 机器人身体 -->
                  <rect x="8" y="14" width="8" height="6" rx="1" fill="#60A5FA"/>
                  <!-- 机器人手臂 -->
                  <rect x="5" y="15" width="2" height="3" rx="1" fill="#60A5FA"/>
                  <rect x="17" y="15" width="2" height="3" rx="1" fill="#60A5FA"/>
                  <!-- 机器人腿 -->
                  <rect x="9" y="20" width="2" height="2" rx="1" fill="#818CF8"/>
                  <rect x="13" y="20" width="2" height="2" rx="1" fill="#818CF8"/>
                  <!-- 天线 -->
                  <line x1="10" y1="6" x2="10" y2="4" stroke="#818CF8" stroke-width="1"/>
                  <line x1="14" y1="6" x2="14" y2="4" stroke="#818CF8" stroke-width="1"/>
                  <circle cx="10" cy="4" r="0.5" fill="#818CF8"/>
                  <circle cx="14" cy="4" r="0.5" fill="#818CF8"/>
                </svg>
              </span>
              <span class="font-bold text-lg mb-1">{{ t.features_oem }}</span>
              <span class="text-gray-700 text-base">{{ t.features_oem_desc }}</span>
            </div>
          </n-card>
          
          <!-- Auto top-up -->
          <n-card class="bg-gradient-to-br from-blue-100 via-indigo-100 to-white hover:scale-105 transition-transform">
            <div class="flex flex-col items-center">
              <span class="text-3xl mb-2">⚡</span>
              <span class="font-bold text-lg mb-1">{{ t.features_auto }}</span>
              <span class="text-gray-700 text-base">{{ t.features_auto_desc }}</span>
            </div>
          </n-card>
          
          <!-- Official recharge -->
          <n-card class="bg-gradient-to-br from-blue-100 via-indigo-100 to-white hover:scale-105 transition-transform">
            <div class="flex flex-col items-center">
              <span class="mb-2">
                <svg width="42" height="42" viewBox="0 0 42 42" fill="none">
                  <path d="M21 6C21 6 9 10.2 9 14.4V24.6C9 31.2 21 36 21 36C21 36 33 31.2 33 24.6V14.4C33 10.2 21 6 21 6Z" fill="#22c55e"/>
                  <path d="M16.7 23.6l4 4 7-7" stroke="#fff" stroke-width="2.6" stroke-linecap="round" stroke-linejoin="round"/>
                </svg>
              </span>
              <span class="font-bold text-lg mb-1">{{ t.features_official }}</span>
              <span class="text-gray-700 text-base">{{ t.features_official_desc }}</span>
            </div>
          </n-card>
          
          <!-- Data security -->
          <n-card class="bg-gradient-to-br from-blue-100 via-indigo-100 to-white hover:scale-105 transition-transform">
            <div class="flex flex-col items-center">
              <span class="text-3xl mb-2">🔒</span>
              <span class="font-bold text-lg mb-1">{{ t.features_security }}</span>
              <span class="text-gray-700 text-base">{{ t.features_security_desc }}</span>
            </div>
          </n-card>
        </div>
      </section>

      <!-- Steps Section -->
      <section class="bg-gray-50 py-20 px-6 text-center text-gray-900" id="steps">
        <h3 class="text-4xl font-bold mb-10">{{ t.steps_title }}</h3>
        <div class="max-w-5xl mx-auto grid grid-cols-1 md:grid-cols-4 gap-8">
          <n-card class="hover:scale-105 transition-transform">
            <div class="flex flex-col items-center">
              <div class="text-4xl mb-3">🔑</div>
              <div class="font-bold mb-1">{{ t.step_1_title }}</div>
              <div class="text-gray-700 text-base">{{ t.step_1_desc }}</div>
            </div>
          </n-card>
          
          <n-card class="hover:scale-105 transition-transform">
            <div class="flex flex-col items-center">
              <div class="text-4xl mb-3">📝</div>
              <div class="font-bold mb-1">{{ t.step_2_title }}</div>
              <div class="text-gray-700 text-base">{{ t.step_2_desc }}</div>
            </div>
          </n-card>
          
          <n-card class="hover:scale-105 transition-transform">
            <div class="flex flex-col items-center">
              <div class="mb-3">
                <svg width="40" height="40" fill="none" viewBox="0 0 24 24">
                  <rect x="4" y="9" width="16" height="8" rx="4" fill="#818CF8"/>
                  <rect x="8" y="2" width="8" height="4" rx="2" fill="#60A5FA"/>
                  <rect x="7" y="17" width="10" height="2" rx="1" fill="#60A5FA"/>
                  <circle cx="9.5" cy="13" r="1.2" fill="#fff"/>
                  <circle cx="14.5" cy="13" r="1.2" fill="#fff"/>
                </svg>
              </div>
              <div class="font-bold mb-1">{{ t.step_3_title }}</div>
              <div class="text-gray-700 text-base">{{ t.step_3_desc }}</div>
            </div>
          </n-card>
          
          <n-card class="hover:scale-105 transition-transform">
            <div class="flex flex-col items-center">
              <div class="text-4xl mb-3">✅</div>
              <div class="font-bold mb-1">{{ t.step_4_title }}</div>
              <div class="text-gray-700 text-base">{{ t.step_4_desc }}</div>
            </div>
          </n-card>
        </div>
        
        <div class="max-w-3xl mx-auto mt-14 rounded-2xl overflow-hidden shadow-lg">
          <video controls style="width:100%;background:#000;">
            <source src="https://video-sh.cloudvideocdn.taobao.com/73bbe9f95b148212/c3deacc6f692e04c/20250404_8be5b5f781d3f9ba_514196921671_255715972792866_published_mp4_264_hd_taobao.mp4?auth_key=1748789725-0-0-b826ee773f75a49cf18a84d0b5e6c564&biz=tbs_vsucai-003cb91d33d66930&t=213e018217487870258545496e0e3f&t=213e018217487870258545496e0e3f&b=tbs_vsucai&p=cloudvideo_http_tb_seller_vsucai_publish" type="video/mp4">
            Your browser does not support video.
          </video>
        </div>
      </section>

      <!-- FAQ Section -->
      <section class="py-20 bg-white px-6 text-gray-900" id="faq">
        <h3 class="text-4xl font-bold text-center mb-10">{{ t.faq_title }}</h3>
        <div class="max-w-4xl mx-auto space-y-4 text-lg">
          
          <div class="border-b">
            <div class="py-3 font-semibold flex items-center justify-between cursor-pointer" @click="toggleFaq(0)">
              <span>{{ t.faq_q1 }}</span>
              <span class="text-xl">{{ faqOpenStates[0] ? '−' : '+' }}</span>
            </div>
            <n-collapse-transition>
              <div v-show="faqOpenStates[0]" class="px-2 pb-3 text-gray-700">
                {{ t.faq_a1 }}
              </div>
            </n-collapse-transition>
          </div>
          
          <div class="border-b">
            <div class="py-3 font-semibold flex items-center justify-between cursor-pointer" @click="toggleFaq(1)">
              <span>{{ t.faq_q2 }}</span>
              <span class="text-xl">{{ faqOpenStates[1] ? '−' : '+' }}</span>
            </div>
            <n-collapse-transition>
              <div v-show="faqOpenStates[1]" class="px-2 pb-3 text-gray-700">
                {{ t.faq_a2 }}
              </div>
            </n-collapse-transition>
          </div>
          
          <div class="border-b">
            <div class="py-3 font-semibold flex items-center justify-between cursor-pointer" @click="toggleFaq(2)">
              <span>{{ t.faq_q3 }}</span>
              <span class="text-xl">{{ faqOpenStates[2] ? '−' : '+' }}</span>
            </div>
            <n-collapse-transition>
              <div v-show="faqOpenStates[2]" class="px-2 pb-3 text-gray-700">
                {{ t.faq_a3 }}
              </div>
            </n-collapse-transition>
          </div>
          
          <div class="border-b">
            <div class="py-3 font-semibold flex items-center justify-between cursor-pointer" @click="toggleFaq(3)">
              <span>{{ t.faq_q4 }}</span>
              <span class="text-xl">{{ faqOpenStates[3] ? '−' : '+' }}</span>
            </div>
            <n-collapse-transition>
              <div v-show="faqOpenStates[3]" class="px-2 pb-3 text-gray-700">
                {{ t.faq_a4 }}
              </div>
            </n-collapse-transition>
          </div>
          
        </div>
      </section>

      <!-- Footer -->
      <footer class="bg-gray-100 text-center text-sm p-6 text-gray-600 flex flex-col sm:flex-row justify-center items-center gap-3">
        <span>{{ t.footer_copyright }}</span>
        <span class="hidden sm:inline-block mx-2">|</span>
        <n-button text @click="showTechModal = true" class="underline hover:text-indigo-600">
          {{ t.footer_tech }}
        </n-button>
        <span class="hidden sm:inline-block mx-2">|</span>
        <n-button text @click="showPrivacyModal = true" class="underline hover:text-indigo-600">
          {{ t.footer_privacy }}
        </n-button>
      </footer>

      <!-- Tech Modal -->
      <n-modal v-model:show="showTechModal">
        <n-card style="width: 600px" title="Tech Assurance" :bordered="false" size="huge" role="dialog" aria-modal="true">
          <div class="space-y-4">
            <p>Our platform provides comprehensive technical assurance:</p>
            <ul class="list-disc pl-6 space-y-2">
              <li>24/7 automated system monitoring</li>
              <li>99.9% uptime guarantee</li>
              <li>Secure token handling with encryption</li>
              <li>Real-time transaction processing</li>
              <li>Professional technical support</li>
            </ul>
          </div>
        </n-card>
      </n-modal>

      <!-- Privacy Modal -->
      <n-modal v-model:show="showPrivacyModal">
        <n-card style="width: 600px" title="Privacy Policy" :bordered="false" size="huge" role="dialog" aria-modal="true">
          <div class="space-y-4">
            <p>We are committed to protecting your privacy:</p>
            <ul class="list-disc pl-6 space-y-2">
              <li>We only collect necessary token information</li>
              <li>No personal account credentials are stored</li>
              <li>All data is encrypted in transit and at rest</li>
              <li>Tokens are securely deleted after processing</li>
              <li>We do not share data with third parties</li>
            </ul>
          </div>
        </n-card>
      </n-modal>
      
    </div>
  </div>
</template>

<style scoped>
@keyframes fadeIn {
  0% { opacity: 0; transform: translateY(20px); }
  100% { opacity: 1; transform: translateY(0); }
}

.nav-link {
  color: #111; 
  font-weight: 500; 
  font-size: 1rem; 
  padding: 0.2rem 1rem;
  border-radius: 9999px; 
  border: 2px solid #111; 
  background: transparent;
  transition: border-color 0.18s, color 0.18s, background 0.15s;
  margin: 0 0.05rem;
}

.nav-link:hover, .nav-link:focus {
  border-color: #60a5fa;
  color: #3b82f6;
  background: #f1f6fd;
}

@media (max-width: 600px) {
  .nav-link { 
    font-size: 0.92rem; 
    padding: 0.14rem 0.5rem;
  }
}

.lang-dropdown-btn {
  border: 2px solid #111 !important;
  background: #fff !important;
  color: #111 !important;
  font-weight: 500;
  transition: border-color 0.18s, color 0.18s, background 0.15s;
}

.lang-dropdown-btn:hover {
  border-color: #60a5fa !important;
  color: #3b82f6 !important;
  background: #f1f6fd !important;
}

body { 
  font-family: 'Segoe UI', 'PingFang SC', 'Microsoft YaHei', sans-serif; 
}
</style>