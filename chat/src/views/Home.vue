<script setup lang="ts">
import { ref, onMounted, nextTick, computed } from 'vue'
import { 
  NButton, 
  NCard, 
  NCollapseTransition, 
  NModal 
} from 'naive-ui'
import Header from '@/components/Layout/Header.vue'
import '@/styles/custom.css'

// Language management
const currentLang = ref('zh')
const isLoading = ref(true)

// FAQ state
const faqOpenStates = ref([false, false, false, false])

// Import language files
import enLang from '@/lang/en'
import zhLang from '@/lang/zh'
import ruLang from '@/lang/ru'

// Language dictionary
const langDict: Record<string, any> = {
  en: enLang,
  zh: zhLang,
  ru: ruLang
}

// Modal states
const showTechModal = ref(false)
const showPrivacyModal = ref(false)

// Language detection and management
const detectLang = () => {
  const lang = navigator.language || (navigator as any).userLanguage
  if (lang.startsWith('zh')) return 'zh'
  if (lang.startsWith('ru')) return 'ru'
  return 'en'
}

const toggleFaq = (index: number) => {
  faqOpenStates.value = faqOpenStates.value.map((state, i) => 
    i === index ? !state : false
  )
}

const initParticles = () => {
  if (typeof window !== 'undefined' && (window as any).particlesJS) {
    (window as any).particlesJS("particles-js", {
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

// Lifecycle
onMounted(async () => {
  // Initialize language
  const savedLang = localStorage.getItem('lang') || detectLang()
  currentLang.value = savedLang
  
  // Set page title
  document.title = langDict[savedLang].page_title
  
  // Load particles.js
  if (typeof window !== 'undefined') {
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
    <div class="bg-gradient-to-br from-black via-gray-800 to-gray-100 text-gray-900 min-h-screen animate-[fadeIn_1s_ease-in-out]">
      
      <!-- Header 组件 -->
      <Header />

      <!-- Hero Section -->
      <section class="py-20 text-center px-4 relative z-10">
        <img 
          alt="ChatGPT Logo" 
          class="w-16 h-16 mx-auto rounded-full animate-spin-slow" 
          src="@/assets/images/ChatGPT-Logo.svg"
        />
        <h2 class="text-4xl sm:text-5xl font-bold mt-5 mb-4 text-white drop-shadow">
          {{ t.hero_title }}
        </h2>
        <p class="text-xl mb-8 text-white drop-shadow">
          {{ t.hero_sub }}
        </p>
        <div class="flex justify-center gap-3 items-center flex-wrap">
          <n-button 
            type="primary" 
            size="large"
            class="bg-white text-black font-semibold px-8 py-3 rounded-full shadow-lg hover:shadow-2xl hover:scale-105 transition-all text-lg hover:bg-black hover:text-white"
            tag="a"
            @click="$router.push('/rooms')"
          >
            {{ t.hero_btn }}
          </n-button>
        </div>
      </section>

      <!-- Features Section -->
      <section class="py-20 bg-white text-gray-900 text-center px-6" id="features">
        <h3 class="text-4xl font-bold mb-10">{{ t.features_title }}</h3>
        <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-8 max-w-6xl mx-auto">
          <!-- OEM - 无人值守 -->
          <n-card class="feature-card">
            <div class="flex flex-col items-center">
              <span class="text-3xl mb-2">🤖</span>
              <span class="font-bold text-lg mb-1">{{ t.features_oem }}</span>
              <span class="text-gray-700 text-base">{{ t.features_oem_desc }}</span>
            </div>
          </n-card>
          
          <!-- Auto top-up -->
          <n-card class="feature-card">
            <div class="flex flex-col items-center">
              <span class="text-3xl mb-2">⚡</span>
              <span class="font-bold text-lg mb-1">{{ t.features_auto }}</span>
              <span class="text-gray-700 text-base">{{ t.features_auto_desc }}</span>
            </div>
          </n-card>
          
          <!-- Official recharge -->
          <n-card class="feature-card">
            <div class="flex flex-col items-center">
              <span class="text-3xl mb-2">✅️</span>
              <span class="font-bold text-lg mb-1">{{ t.features_official }}</span>
              <span class="text-gray-700 text-base">{{ t.features_official_desc }}</span>
            </div>
          </n-card>
          
          <!-- Data security -->
          <n-card class="feature-card">
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
          <n-card class="feature-card">
            <div class="flex flex-col items-center">
              <div class="text-4xl mb-3">🔑</div>
              <div class="font-bold mb-1">{{ t.step_1_title }}</div>
              <div class="text-gray-700 text-base">{{ t.step_1_desc }}</div>
            </div>
          </n-card>
          
          <n-card class="feature-card">
            <div class="flex flex-col items-center">
              <div class="text-4xl mb-3">📝</div>
              <div class="font-bold mb-1">{{ t.step_2_title }}</div>
              <div class="text-gray-700 text-base">{{ t.step_2_desc }}</div>
            </div>
          </n-card>
          
          <n-card class="feature-card">
            <div class="flex flex-col items-center">
              <div class="text-4xl mb-3">🤖</div>
              <div class="font-bold mb-1">{{ t.step_3_title }}</div>
              <div class="text-gray-700 text-base">{{ t.step_3_desc }}</div>
            </div>
          </n-card>
          
          <n-card class="feature-card">
            <div class="flex flex-col items-center">
              <div class="text-4xl mb-3">✅</div>
              <div class="font-bold mb-1">{{ t.step_4_title }}</div>
              <div class="text-gray-700 text-base">{{ t.step_4_desc }}</div>
            </div>
          </n-card>
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
        <n-button text @click="showTechModal = true" class="underline hover:text-black">
          {{ t.footer_tech }}
        </n-button>
        <span class="hidden sm:inline-block mx-2">|</span>
        <n-button text @click="showPrivacyModal = true" class="underline hover:text-black">
          {{ t.footer_privacy }}
        </n-button>
      </footer>

      <!-- Tech Modal -->
      <n-modal v-model:show="showTechModal">
        <n-card style="width: 600px; max-width: 90vw;" title="技术保障" :bordered="false" size="huge" role="dialog" aria-modal="true">
          <div class="space-y-4">
            <p>我们的平台提供全面的技术保障：</p>
            <ul class="list-disc pl-6 space-y-2">
              <li>7x24 小时系统监控和维护</li>
              <li>99.9% 服务可用性保证</li>
              <li>多 AI 模型无缝切换</li>
              <li>智能负载均衡，确保最佳响应速度</li>
              <li>专业技术团队实时支持</li>
              <li>定期更新和优化 AI 模型</li>
            </ul>
          </div>
        </n-card>
      </n-modal>

      <!-- Privacy Modal -->
      <n-modal v-model:show="showPrivacyModal">
        <n-card style="width: 600px; max-width: 90vw;" title="隐私条款" :bordered="false" size="huge" role="dialog" aria-modal="true">
          <div class="space-y-4">
            <p>我们承诺保护您的隐私安全：</p>
            <ul class="list-disc pl-6 space-y-2">
              <li>采用端到端加密技术保护所有对话数据</li>
              <li>不会存储或分享您的个人敏感信息</li>
              <li>对话记录仅保存在您的账户中</li>
              <li>您可以随时删除自己的对话历史</li>
              <li>严格遵守数据保护法规和隐私政策</li>
              <li>不会将您的数据用于第三方用途</li>
            </ul>
          </div>
        </n-card>
      </n-modal>
      
    </div>
  </div>
</template>
