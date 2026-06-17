import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom'
import Landing from './pages/Landing'
import Login from './pages/Login'
import Register from './pages/Register'
import Pay from './pages/Pay'
import Layout from './components/Layout'
import ProtectedRoute from './components/ProtectedRoute'
import Overview from './pages/dashboard/Overview'
import TelegramAccounts from './pages/dashboard/TelegramAccounts'
import Cards from './pages/dashboard/Cards'
import Webhook from './pages/dashboard/Webhook'
import CreateTransaction from './pages/dashboard/CreateTransaction'
import Transactions from './pages/dashboard/Transactions'
import WebhookLogs from './pages/dashboard/WebhookLogs'
import AiIntegration from './pages/dashboard/AiIntegration'

export default function App() {
  return (
    <BrowserRouter>
      <Routes>
        <Route path="/" element={<Landing />} />
        <Route path="/login" element={<Login />} />
        <Route path="/register" element={<Register />} />
        <Route path="/pay/:id" element={<Pay />} />
        <Route path="/dashboard" element={<ProtectedRoute><Layout /></ProtectedRoute>}>
          <Route index element={<Overview />} />
          <Route path="telegram" element={<TelegramAccounts />} />
          <Route path="cards" element={<Cards />} />
          <Route path="webhook" element={<Webhook />} />
          <Route path="webhook-logs" element={<WebhookLogs />} />
          <Route path="transaction" element={<CreateTransaction />} />
          <Route path="transactions" element={<Transactions />} />
          <Route path="ai" element={<AiIntegration />} />
        </Route>
        <Route path="*" element={<Navigate to="/" replace />} />
      </Routes>
    </BrowserRouter>
  )
}
