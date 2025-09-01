<script setup lang="ts">
import { ref, onMounted, nextTick, h, Component } from 'vue'
// Use flag SVGs from public directory
const cnFlag = '/flags/CN.svg'
const usFlag = '/flags/US.svg'
const ruFlag = '/flags/RU.svg'


// Language management
const currentLang = ref('zh')
const isLoading = ref(true)

// FAQ state
const faqOpenStates = ref([false, false, false, false])

// Import language files
import enLang from '../lang/en.js'
import zhLang from '../lang/zh.js'
import ruLang from '../lang/ru.js'

// Import component styles
import '../assets/css/index.css'

// Language dictionary
const langDict = {
  en: enLang,
  zh: zhLang,
  ru: ruLang
}

// Modal states
const showTechModal = ref(false)
const showPrivacyModal = ref(false)

// Language detection and management
const detectLang = () => {
  const lang = navigator.language || navigator.userLanguage
  if (lang.startsWith('zh')) return 'zh'
  if (lang.startsWith('ru')) return 'ru'
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

// Helper function to render flag icons
const renderIcon = (flagSvg: string) => {
  return () => h('img', { 
    src: flagSvg, 
    style: 'width: 20px; height: 15px; border-radius: 2px;',
    alt: 'Flag'
  })
}

// Language options for dropdown
const langOptions = [
  {
    label: 'English',
    key: 'en',
    icon: renderIcon(usFlag)
  },
  {
    label: '中文',
    key: 'zh',
    icon: renderIcon(cnFlag)
  },
  {
    label: 'Русский',
    key: 'ru',
    icon: renderIcon(ruFlag)
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
    <div class="bg-gradient-to-br from-black via-gray-800 to-gray-100 text-gray-900 min-h-screen animate-[fadeIn_1s_ease-in-out]">
      
      <!-- Header -->
      <header class="bg-white p-4 sticky top-0 z-40 shadow-xl border-b border-black/20">
        <div class="max-w-7xl mx-auto flex justify-between items-center flex-col sm:flex-row gap-3 sm:gap-0">
          <div>
            <button 
              @click="scrollToTop" 
              class="nav-link-active"
            >
              {{ t.nav_home }}
            </button>
          </div>
          <nav class="flex gap-2 items-center justify-center">
            <a href="" class="nav-link">
              {{ t.hero_btn }}
            </a>
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
                <span v-if="currentLang === 'zh'" class="flex items-center gap-1">
                  <img :src="cnFlag" alt="China Flag" style="width: 20px; height: 15px; border-radius: 2px;" />
                  中文
                </span>
                <span v-else-if="currentLang === 'ru'" class="flex items-center gap-1">
                  <img :src="ruFlag" alt="Russia Flag" style="width: 20px; height: 15px; border-radius: 2px;" />
                  Русский
                </span>
                <span v-else class="flex items-center gap-1">
                  <img :src="usFlag" alt="US Flag" style="width: 20px; height: 15px; border-radius: 2px;" />
                  English
                </span>
              </n-button>
            </n-dropdown>
          </nav>
        </div>
      </header>

      <!-- Hero Section -->
      <section class="py-20 text-center px-4 relative z-10">
        <img 
          alt="ChatGPT Logo" 
          class="w-16 h-16 mx-auto rounded-full animate-spin-slow" 
          src="../assets/images/ChatGPT-Logo.svg"
        />
        <h2 class="text-4xl sm:text-5xl font-bold  mt-5 mb-4 text-white drop-shadow">
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
            href="https://www.ow520.com/"
            target="_blank"
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

