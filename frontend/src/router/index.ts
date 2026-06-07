import { createRouter, createWebHistory, type RouteRecordRaw } from 'vue-router'
import AuthLayout from '@/layouts/AuthLayout.vue'
import AppLayout from '@/layouts/AppLayout.vue'
import ReaderLayout from '@/layouts/ReaderLayout.vue'
import AdminLayout from '@/layouts/AdminLayout.vue'
import LoginView from '@/views/auth/LoginView.vue'
import RegisterView from '@/views/auth/RegisterView.vue'
import BookshelfView from '@/views/bookshelf/BookshelfView.vue'
import BookshelfDetailView from '@/views/bookshelf/BookshelfDetailView.vue'
import LibraryView from '@/views/library/LibraryView.vue'
import LibraryDetailView from '@/views/library/LibraryDetailView.vue'
import MyLibraryView from '@/views/library/MyLibraryView.vue'
import ReaderView from '@/views/reader/ReaderView.vue'
import UploadView from '@/views/upload/UploadView.vue'
import SettingsView from '@/views/settings/SettingsView.vue'
import AdminHomeView from '@/views/admin/AdminHomeView.vue'
import AdminUsersView from '@/views/admin/AdminUsersView.vue'
import AdminLibraryView from '@/views/admin/AdminLibraryView.vue'
import AdminCategoriesView from '@/views/admin/AdminCategoriesView.vue'
import AdminTagsView from '@/views/admin/AdminTagsView.vue'
import AdminSettingsView from '@/views/admin/AdminSettingsView.vue'
import NotFoundView from '@/views/NotFoundView.vue'
import { installRouterGuards } from './guards'

const routes: RouteRecordRaw[] = [
  { path: '/', redirect: '/bookshelf' },
  {
    path: '/',
    component: AuthLayout,
    children: [
      { path: 'login', name: 'login', component: LoginView },
      { path: 'register', name: 'register', component: RegisterView }
    ]
  },
  {
    path: '/',
    component: AppLayout,
    meta: { requiresAuth: true },
    children: [
      { path: 'bookshelf', name: 'bookshelf', component: BookshelfView },
      { path: 'bookshelf/:id', name: 'bookshelf-detail', component: BookshelfDetailView, props: true },
      { path: 'library', name: 'library', component: LibraryView },
      { path: 'library/mine', name: 'library-mine', component: MyLibraryView },
      { path: 'library/:id', name: 'library-detail', component: LibraryDetailView, props: true },
      { path: 'upload', name: 'upload', component: UploadView },
      { path: 'settings', name: 'settings', component: SettingsView }
    ]
  },
  {
    path: '/reader/:bookId',
    component: ReaderLayout,
    meta: { requiresAuth: true },
    children: [{ path: '', name: 'reader', component: ReaderView, props: true }]
  },
  {
    path: '/admin',
    component: AdminLayout,
    meta: { requiresAuth: true, requiresAdmin: true },
    children: [
      { path: '', name: 'admin', component: AdminHomeView },
      { path: 'users', name: 'admin-users', component: AdminUsersView },
      { path: 'books', name: 'admin-books', component: AdminLibraryView },
      { path: 'library', name: 'admin-library', component: AdminLibraryView },
      { path: 'categories', name: 'admin-categories', component: AdminCategoriesView },
      { path: 'tags', name: 'admin-tags', component: AdminTagsView },
      { path: 'settings', name: 'admin-settings', component: AdminSettingsView }
    ]
  },
  { path: '/:pathMatch(.*)*', name: 'not-found', component: NotFoundView }
]

const router = createRouter({
  history: createWebHistory(),
  routes,
  scrollBehavior: () => ({ top: 0 })
})

installRouterGuards(router)

export default router
