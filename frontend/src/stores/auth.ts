import { defineStore } from "pinia";
import { ref, computed } from "vue";
import { jwtDecode } from "jwt-decode";
import { authService } from "@/service/auth";

function isTokenExpired(token: string | null) {
  if (!token) return true;
  try {
    const payload = (jwtDecode as any)(token);
    if (!payload || typeof payload === "string") return true;
    if (!payload.exp) return false;
    return Date.now() >= payload.exp * 1000;
  } catch (e) {
    return true;
  }
}

export const useAuthStore = defineStore("auth", () => {
  const accessToken = ref<string | null>(null);
  const user_id = ref<string | null>(null);
  const username = ref<string | null>(null);

  const isAuthenticated = computed(
    () => !!accessToken.value && !isTokenExpired(accessToken.value)
  );
  const authHeader = computed(() =>
    accessToken.value ? { Authorization: "Bearer " + accessToken.value } : {}
  );

  function setToken(token: string | null) {
    // Save token in localStorage and in state
    accessToken.value = token;
    try {
      if (token) {
        localStorage.setItem("access_token", token);
      } else {
        localStorage.removeItem("access_token");
      }
    } catch (e) {
      console.error("Failed to access localStorage", e);
    }
    if (token) {
      try {
        const payload = (jwtDecode as any)(token);
        user_id.value = payload?.user_id ?? null;
        username.value = payload?.username ?? null;
      } catch (e) {
        console.log("Failed to decode token", e);
        user_id.value = null;
        username.value = null;
      }
    } else {
      user_id.value = null;
      username.value = null;
    }
  }

  function logout() {
    accessToken.value = null;
    user_id.value = null;
    username.value = null;
    try {
      localStorage.removeItem("access_token");
    } catch (e) {
      console.error("Failed to access localStorage", e);
    }
  }

  async function login(username: string, password: string) {
    try {
      const login_response = await authService.login(username, password);

      if (login_response.error) {
        throw new Error(login_response.error);
      }
      console.log(login_response.access_token);
      setToken(login_response.access_token);
    } catch (e: any) {
      // Normalize any error (axios or otherwise) into an Error with a friendly message
      const msg = e?.response?.data?.detail ?? e?.response?.data?.message ?? e?.message ?? String(e)
      throw new Error(msg)
    }
  }

  // Restore token from localStorage if present
  try {
    const stored = localStorage.getItem("access_token");
    if (stored) setToken(stored);
  } catch (e) {
    console.error("Failed to access localStorage", e);
  }

  return {
    accessToken,
    user_id,
    username,
    isAuthenticated,
    authHeader,
    setToken,
    logout,
    login,
  };
});
