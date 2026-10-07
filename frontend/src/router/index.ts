import { createRouter, createWebHashHistory } from 'vue-router'
import { useAuth } from '@/stores/auth'

const routes = [
  { path: '/login', name: 'login', component: () => import('@/pages/Login.vue'), meta: { publik: true, polos: true } },
  {
    path: '/',
    component: () => import('@/layouts/AppLayout.vue'),
    children: [
      { path: '', name: 'dashboard', component: () => import('@/pages/Dashboard.vue'), meta: { judul: 'Dashboard' } },
      { path: 'library', name: 'library', component: () => import('@/pages/CpLibrary.vue'), meta: { judul: 'CP Library' } },
      { path: 'library/:kode', name: 'cp-detail', component: () => import('@/pages/CpDetail.vue'), meta: { judul: 'Detail CP' } },
      { path: 'approval', name: 'approval', component: () => import('@/pages/Approval.vue'), meta: { judul: 'Approval' } },
      { path: 'cp-aktif', name: 'cp-aktif', component: () => import('@/pages/CpAktif.vue'), meta: { judul: 'CP Aktif' } },
      { path: 'evaluasi', name: 'evaluasi', component: () => import('@/pages/Evaluasi.vue'), meta: { judul: 'Evaluasi CP' } },
      { path: 'dokumen', name: 'dokumen', component: () => import('@/pages/Dokumen.vue'), meta: { judul: 'Dokumen Panduan' } },
      { path: 'padanan/kptl', name: 'kptl', component: () => import('@/pages/Padanan.vue'), meta: { judul: 'Padanan KPTL' } },
      { path: 'padanan/snomed', name: 'snomed', component: () => import('@/pages/Padanan.vue'), meta: { judul: 'Padanan SNOMED-CT' } },
      { path: 'pengaturan', name: 'pengaturan', component: () => import('@/pages/Pengaturan.vue'), meta: { judul: 'Pengaturan', admin: true } },
    ],
  },
  { path: '/:pathMatch(.*)*', redirect: '/' },
]

export const router = createRouter({
  history: createWebHashHistory(),
  routes,
  scrollBehavior: () => ({ top: 0 }),
})

router.beforeEach(async (to) => {
  const auth = useAuth()
  if (!auth.siap) await auth.muatProfil()

  if (!to.meta.publik && !auth.masuk) {
    return { name: 'login', query: to.fullPath !== '/' ? { ke: to.fullPath } : undefined }
  }
  if (to.name === 'login' && auth.masuk) return { path: '/' }
  if (to.meta.admin && !auth.isAdmin) return { path: '/' }
})
