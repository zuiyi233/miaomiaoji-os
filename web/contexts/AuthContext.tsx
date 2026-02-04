import React, { createContext, useContext, useEffect, useMemo, useState, ReactNode } from 'react';
import { User, SystemConfig, RedemptionCode, CodeGenerationConfig } from '../types';
import { apiRequest, getToken, setToken } from '../services/apiClient';
import { loginApi, logoutApi, registerApi } from '../services/authApi';
import { fetchProfileApi } from '../services/userApi';
import { batchUpdateCodesApi, fetchCodesApi, generateCodesApi, getExportCodesUrl, redeemCodeApi } from '../services/redemptionApi';

interface AuthContextType {
  user: User | null;
  login: (username: string, password: string) => Promise<boolean>;
  signup: (
    username: string,
    password: string,
    inviteCode: string
  ) => Promise<{ success: boolean; message?: string }>;
  logout: () => void;
  isLoading: boolean;

  // 兼容模板 UI：保留这些字段/方法，改为最小实现
  allUsers: User[];
  redemptionCodes: RedemptionCode[];
  systemConfig: SystemConfig;
  deleteUser: (userId: string) => void;
  batchGenerateCodes: (
    config: CodeGenerationConfig & { charType?: 'alphanum' | 'num' | 'alpha'; creatorOverride?: string }
  ) => string[];
  batchUpdateCodes: (codes: string[], action: 'delete' | 'disable' | 'enable' | 'renew', value?: any) => void;
  fetchCodes: (page: number, size: number, status: string, search: string, sort: string) => Promise<void>;
  updateSystemConfig: (config: Partial<SystemConfig>) => void;
  generateInviteCode: (days: number, count?: number) => void;
  deleteInviteCode: (code: string) => void;

  performCheckIn: () => { success: boolean; pointsAwarded?: number; streak?: number; message?: string };
  awardPoints: (amount: number, reason?: string) => void;
  exchangePointsForCode: (optionId: string) => { success: boolean; code?: string; message?: string };

  deviceId: string;
  hasAIAccess: boolean;
}

const AuthContext = createContext<AuthContextType | undefined>(undefined);

const DEVICE_ID_KEY = 'nao_device_id_unique';

const DEFAULT_SYSTEM_CONFIG: SystemConfig = {
  checkInBasePointsMax: 20,
  checkInStreakBonus: 5,
  enablePointsExchange: true,
  exchangeOptions: [
    { id: 'opt_basic', name: '30天 AI 体验卡', cost: 1000, durationDays: 30, description: '适合短期尝鲜' },
    { id: 'opt_pro', name: '90天 创作季卡', cost: 2500, durationDays: 90, description: '九折优惠，稳定创作' },
  ],
};

const getOrCreateDeviceId = (): string => {
  let id = localStorage.getItem(DEVICE_ID_KEY);
  if (!id) {
    id = `dev_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`;
    localStorage.setItem(DEVICE_ID_KEY, id);
  }
  return id;
};

const getRequestId = () => `req_${Date.now()}_${Math.random().toString(36).slice(2, 10)}`;

function mapUserFromProfile(profile: any): User {
  const aiAccessUntil = profile.ai_access_until ? Date.parse(profile.ai_access_until) : 0;
  return {
    id: String(profile.id),
    username: profile.username || 'unknown',
    role: profile.role === 'admin' ? 'admin' : 'user',
    createdAt: Date.now(),
    points: Number(profile.points || 0),
    checkInStreak: Number(profile.check_in_streak || 0),
    lastCheckIn: 0,
    aiAccessUntil: aiAccessUntil > 0 ? aiAccessUntil : undefined,
  };
}

export const AuthProvider: React.FC<{ children: ReactNode }> = ({ children }) => {
  const [user, setUser] = useState<User | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [deviceId, setDeviceId] = useState('');

  // 保留模板字段，最小化实现（后端阶段不做兑换码/多用户管理）
  const [allUsers] = useState<User[]>([]);
  const [redemptionCodes, setRedemptionCodes] = useState<RedemptionCode[]>([]);
  const [systemConfig] = useState<SystemConfig>(DEFAULT_SYSTEM_CONFIG);

  useEffect(() => {
    setDeviceId(getOrCreateDeviceId());

    const bootstrap = async () => {
      const token = getToken();
      if (!token) {
        setIsLoading(false);
        return;
      }

      try {
        const profile = await fetchProfileApi();
        setUser(mapUserFromProfile(profile));
        await refreshCodes();
      } catch {
        // token 失效或后端不可用，清理登录态
        setToken(null);
        setUser(null);
      } finally {
        setIsLoading(false);
      }
    };

    bootstrap();
  }, []);

  const hasAIAccess = useMemo(() => {
    if (!user) return false;
    if (user.role === 'admin') return true;
    if (!user.aiAccessUntil) return false;
    return user.aiAccessUntil > Date.now();
  }, [user]);

  const login = async (username: string, password: string): Promise<boolean> => {
    try {
      await loginApi({ username, password });
      const profile = await fetchProfileApi();
      setUser(mapUserFromProfile(profile));
      await refreshCodes();
      // 登录后上报心跳（仅基础信息）
      apiRequest('/api/v1/users/heartbeat', {
        method: 'POST',
        body: JSON.stringify({ device_id: deviceId })
      }).catch(() => {});
      return true;
    } catch {
      return false;
    }
  };

  const signup = async (
    username: string,
    password: string,
    inviteCode: string
  ): Promise<{ success: boolean; message?: string }> => {
    try {
      await registerApi({ username, password });
      const profile = await fetchProfileApi();
      setUser(mapUserFromProfile(profile));
      if (inviteCode) {
        await redeemCodeApi({
          request_id: getRequestId(),
          idempotency_key: `redeem_${deviceId}_${inviteCode}`,
          code: inviteCode,
          device_id: deviceId,
          client_time: new Date().toISOString(),
          app_id: 'novel-agent-os',
          platform: 'web',
          app_version: (import.meta as any).env?.VITE_APP_VERSION || 'dev',
          result_status: 'success',
          entitlement_delta: {
            entitlement_type: 'ai_access',
            grant_mode: 'add_days'
          }
        });
        const refreshed = await fetchProfileApi();
        setUser(mapUserFromProfile(refreshed));
        await refreshCodes();
      }
      return { success: true };
    } catch (e: any) {
      return { success: false, message: e?.message || '注册失败' };
    }
  };

  const logout = () => {
    // 先清理本地，再异步通知后端
    setUser(null);
    setRedemptionCodes([]);
    setToken(null);
    logoutApi().catch(() => {
      // 忽略网络错误
    });
  };

  const refreshCodes = async (page = 1, size = 500, status = 'all', search = '', sort = 'desc') => {
    try {
      if (!user || user.role !== 'admin') {
        setRedemptionCodes([]);
        return;
      }
      const data = await fetchCodesApi(page, size, status, search, sort);
      setRedemptionCodes(
        (data.list || []).map((item) => ({
          code: item.code,
          createdAt: item.created_at ? Date.parse(item.created_at) : Date.now(),
          expiresAt: item.expires_at ? Date.parse(item.expires_at) : Date.now(),
          maxUses: item.max_uses || 1,
          usedCount: item.used_count || 0,
          usedByUserIds: [],
          createdBy: String(item.created_by || ''),
          batchId: item.batch_id,
          tags: item.tags || [],
          note: item.note,
          prefix: item.prefix,
          status: (item.status || 'active') as any,
          source: (item.source || 'admin') as any,
        }))
      );
    } catch {
      setRedemptionCodes([]);
    }
  };

  return (
    <AuthContext.Provider
      value={{
        user,
        login,
        signup,
        logout,
        isLoading,
        deviceId,
        hasAIAccess,

        // 兼容字段：最小实现
        allUsers,
        redemptionCodes,
        systemConfig,
        deleteUser: () => {},
        batchGenerateCodes: (config) => {
          if (!user || user.role !== 'admin') return [];
          generateCodesApi({
            prefix: config.prefix,
            length: config.length,
            count: config.count,
            validity_days: config.validityDays,
            max_uses: config.maxUses,
            tags: config.tags,
            note: config.note,
            source: config.source || 'admin'
          }).then(() => refreshCodes()).catch(() => {});
          return [];
        },
        batchUpdateCodes: (codes, action, value) => {
          if (!user || user.role !== 'admin') return;
          batchUpdateCodesApi({ codes, action, value }).then(() => refreshCodes()).catch(() => {});
        },
        fetchCodes: (page, size, status, search, sort) => refreshCodes(page, size, status, search, sort),
        updateSystemConfig: () => {},
        generateInviteCode: () => {},
        deleteInviteCode: () => {},
        performCheckIn: () => ({ success: false, message: '后端联调阶段暂不支持签到' }),
        awardPoints: () => {},
        exchangePointsForCode: () => ({ success: false, message: '后端联调阶段暂不支持兑换' }),
      }}
    >
      {children}
    </AuthContext.Provider>
  );
};

export const useAuth = () => {
  const context = useContext(AuthContext);
  if (!context) throw new Error('useAuth must be used within an AuthProvider');
  return context;
};
