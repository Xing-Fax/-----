<template>
  <!-- 人机验证对话框 -->
  <div>
    <a-modal v-model:open="open" ok-text="确认" cancel-text="取消" title="人机验证" :confirm-loading="confirmLoading" @ok="handleOk">
      <p>{{ modalText }}</p>
      <div class="captcha-container">
        <!-- 验证码图片 -->
        <img :src="captchaImage" alt="CAPTCHA" class="captcha-image" @click="refreshCaptcha" />
        <!-- 验证码输入框 -->
        <a-input v-model:value="captchaCode" @keypress.enter="handleOk" placeholder="请输入验证码" />
      </div>
    </a-modal>
  </div>
  <!-- 步骤条 -->
  <div class="title-container">
    <a-steps :items="items" style="margin: 0px 15% 0px 15%; margin-bottom: 20px; width: 70%; min-width: 0px;"></a-steps>
  </div>

  <div class="title-container">
    <!-- 第一部分表单 -->
    <a-form v-if="step === 1" :model="formState" name="basic" @finish="onFinishStep1">
      <a-form-item name="username" :rules="[{ required: true, message: '请输入你的用户名!' }]">
        <a-input v-model:value="formState.username" class="input-field" placeholder="用户名、Email或手机号">
          <template #prefix><user-outlined /></template>
        </a-input>
      </a-form-item>

      <a-form-item :wrapper-col="{ offset: 0, span: 16 }">
        <a-button type="primary" html-type="submit" class="submit-button">下一步</a-button>
      </a-form-item>
    </a-form>

    <!-- 第二部分表单 -->
    <a-form v-if="step === 2" :model="formState" name="basic" @finish="onFinishStep2">
      <a-form-item name="contact" :rules="[{ required: true, message: '请选择验证方式' }]">
        <a-select v-model:value="formState.contact" class="input-field" placeholder="请选择验证方式" style="text-align:left;">
          <a-select-option v-for="(contact, index) in contactOptions" :key="index" :value="contact">
            {{ contact }}
          </a-select-option>
        </a-select>
      </a-form-item>
      
      <a-form-item name="extraInput" :rules="[{ required: true, message: '请输入验证码' }]">
        <a-input-group compact>
          <a-input v-model:value="formState.extraInput" style="width: 200px; height: 40px;text-align:left;" placeholder="请输入验证码"/>
          <a-button type="primary" style="height: 40px;" @click="handleButtonClick">发送验证码</a-button>
        </a-input-group>
      </a-form-item>
      
      <a-form-item :wrapper-col="{ offset: 0, span: 16 }">
        <a-button type="primary" html-type="submit" class="submit-button">下一步</a-button>
      </a-form-item>
    </a-form>

    <!-- 第三部分重置密码表单 -->
    <a-form v-if="step === 3" :model="formState" name="resetPasswordForm" @finish="onFinishResetPassword">
      <a-form-item name="newPassword" :rules="[{ validator: checkPassword }]">
        <a-input-password v-model:value="formState.newPassword" class="input-pass" placeholder="请输入新密码" />
      </a-form-item>

      <a-form-item name="confirmPassword" :rules="[{ validator: checkConfirmPassword }]">
        <a-input-password v-model:value="formState.confirmPassword" class="input-pass" placeholder="确认密码" />
      </a-form-item>

      <a-form-item :wrapper-col="{ offset: 0, span: 16 }">
        <a-button type="primary" html-type="submit" class="submit-button">下一步</a-button>
      </a-form-item>
    </a-form>

    <!-- 第三部分重置密码表单 -->
    <a-form v-if="step === 4" name="resetPasswordForm" @finish="onFinishResetSuccess">
      <a-result :status="resultStatus" :title="resultTitle"/>
    </a-form>
  </div>
</template>

<script lang="ts" setup>
import axios from 'axios';
import { ref, reactive, h } from 'vue';
import { message } from 'ant-design-vue';
import { UserOutlined, SolutionOutlined, KeyOutlined } from '@ant-design/icons-vue';
import JSEncrypt from 'jsencrypt';

/** 验证码及弹窗相关变量 */
const open = ref(false);
const modalText = ref('请输入图中验证码');
const captchaImage = ref('');
const captchaId = ref('');
const captchaCode = ref('');
const confirmLoading = ref(false);

/** 表单状态管理 */
const step = ref(1);
const formState = reactive({
  username: '',
  contact: '',
  extraInput: '',
  newPassword: '',
  confirmPassword: '',
});

/** 结果页相关变量 */
const resultStatus = ref('error');
const resultTitle = ref('操作成功');

/** 验证方式选项 */
const contactOptions = ref<string[]>([]);

/** 步骤条配置 */
const items = reactive([
  { title: '账号', status: 'process', icon: h(UserOutlined) },
  { title: '验证', status: 'wait', icon: h(SolutionOutlined) },
  { title: '重置', status: 'wait', icon: h(KeyOutlined) },
]);

async function encryptPassword(password: string): Promise<string> {
  // 获取后端公钥
  const response = await fetch("/api/public-key");
  const data = await response.json();
  const publicKey = data.publicKey;

  // 使用公钥加密密码
  const encryptor = new JSEncrypt();
  encryptor.setPublicKey(publicKey);
  const encrypted = encryptor.encrypt(password);
  if (!encrypted) {
    throw new Error("Failed to encrypt password");
  }
  return encrypted;
}

function updateStatus(stepIndex: number, newStatus: string) {
  if (stepIndex >= 0 && stepIndex < items.length) {
    items[stepIndex].status = newStatus;
  } else {
    console.error('步骤索引超出范围');
  }
}

function errorInfo(code: string) {
  const errorMessages: Record<string, string> = {
    "10001": "生成图形验证码失败",
    "10002": "解密密码失败",
    "10003": "二次校验密码不通过",
    "10004": "LDAP密码重置失败",
    "10005": "无效的验证方式",
    "10006": "无效的验证码",
    "10007": "LDAP初始化失败",
    "10008": "获取用户信息失败",
    "10009": "验证码发送间隔太短",
    "10010": "邮箱验证码发送失败",
    "10011": "手机号验证码发送失败",
    "10012": "未查找到用户信息",
  };

  const messageText = errorMessages[code];
  if (messageText) {
    message.error(messageText);
  } else {
    message.error("未知错误");
  }
}

// 第一步表单提交, 查找用户的邮箱和手机号
const onFinishStep1 = async (values: any) => {
  try {
    const formData = new FormData();
    formData.append("username", values.username);

    const response = await fetch('/api/get-user-info', {
      method: 'POST',
      body: formData,
    });

    if (!response.ok) {
      throw new Error(`HTTP error! status: ${response.status}`);
    }

    const data = await response.json();

    if (data.code == 200) {
      // 成功, 列出手机号和邮箱
      contactOptions.value = [data.mobile, data.mail];
      formState.contact = contactOptions.value[0];
      updateStatus(0, 'finish');
      updateStatus(1, 'process');
      step.value = 2;
    } else {
      errorInfo(data.code)
    }
  } catch (error) {
    message.error("请求失败" + error);
  }
};

// 请求获取验证码图片和ID
const fetchCaptcha = async () => {
  try {
    const response = await axios.get('/api/generate-captcha');
    if (response.data && response.data.id && response.data.url) {
      captchaId.value = response.data.id; // 保存验证码ID
      captchaImage.value = response.data.url; // 设置验证码图片URL
    } else {
      errorInfo(response.data.code)
      message.error("获取验证码失败，请稍后再试");
    }
  } catch (error) {
    message.error("请求验证码失败" + error);
  }
};

// 第二步发送验证码前人机验证

// 发送验证码，先人机验证
const handleButtonClick = () => {
  // 显示对话框
  open.value = true;
  // 清空输入
  captchaCode.value = '';
  fetchCaptcha()

};

// 刷新验证码
const refreshCaptcha = () => {
  captchaCode.value = '';
  fetchCaptcha()
}

// 判断是否是邮箱
const isEmail = (str: string): boolean => str.includes('@');

// 判断是否是手机号
const isPhoneNumber = (str: string): boolean => !str.includes('@');

const handleOk = async () => {
  try {
    // 检查非空
    if (!captchaCode.value) {
      message.error("请输入验证码");
      return;
    }
    confirmLoading.value = true;
    const formData = new FormData();
    // 验证类型
    if (isPhoneNumber(formState.contact)) {
      formData.append("type", "mobile");
    } else if (isEmail(formState.contact)) {
      formData.append("type", "mail");
    }
    // 用户名称
    formData.append("username", formState.username);
    // 验证码ID
    formData.append("verifyID", captchaId.value);
    // 验证码值
    formData.append("verifyCode", captchaCode.value);

    const response = await fetch('/api/send-code', {
      method: 'POST',
      body: formData,
    });

    if (!response.ok) {
      throw new Error(`HTTP error! status: ${response.status}`);
    }

    const data = await response.json();

    if (data.code == 200) {
      // 成功
      message.success("验证码已发送")
      confirmLoading.value = false;
      open.value = false;
    } else {
      errorInfo(data.code)
      captchaCode.value = '';
      fetchCaptcha()
      confirmLoading.value = false;
    }
  } catch (error) {
    message.error("验证失败" + error);
    confirmLoading.value = false;
  }
};

// 验证码验证
const onFinishStep2 = async () => {
  try {
    const formData = new FormData();
    formData.append("username", formState.username);
    formData.append("verifyCode", formState.extraInput);
    // 验证类型
    if (isPhoneNumber(formState.contact)) {
      formData.append("type", "mobile");
    } else if (isEmail(formState.contact)) {
      formData.append("type", "mail");
    } 

    const response = await fetch('/api/verification-code', {
      method: 'POST',
      body: formData,
    });

    if (!response.ok) {
      throw new Error(`HTTP error! status: ${response.status}`);
    }

    const data = await response.json();

    if (data.code == 200) {
      message.success("验证成功")
      step.value = 3;
      updateStatus(1, 'finish');
      updateStatus(2, 'process');
    } else {
      errorInfo(data.code)
    }
  } catch (error) {
    message.error("验证失败" + error);
  }
};

// 第三步, 输入新密码, 提交修改
const onFinishResetPassword = async () => {
  try {
    const formData = new FormData();
    formData.append("username", formState.username);
    formData.append("verifyCode", formState.extraInput);
    const encryptedPassword = await encryptPassword(formState.confirmPassword);
    formData.append("newPassword", encryptedPassword);
    // 验证类型
    if (isPhoneNumber(formState.contact)) {
      formData.append("type", "mobile");
    } else if (isEmail(formState.contact)) {
      formData.append("type", "mail");
    } 

    const response = await fetch('/api/reset-password', {
      method: 'POST',
      body: formData,
    });

    if (!response.ok) {
      throw new Error(`HTTP error! status: ${response.status}`);
    }

    const data = await response.json();

    if (data.code == 200) {
      message.success("重置成功(✿◠‿◠)")
      resultStatus.value = "success";
      resultTitle.value = "重置成功";
      step.value = 4;
    } else {
      errorInfo(data.code)
      resultStatus.value = "error";
      resultTitle.value = "操作失败";
    }
  } catch (error) {
    message.error("重置失败" + error);
  }
};

const onFinishResetSuccess = () => {
  
}

const checkPassword = (_rule: any, value: string, callback: Function) => {
  const passwordRegEx = /^(?=.*[A-Za-z])(?=.*\d)[A-Za-z\d!@#$%^&*(),.?":{}|<>]{8,}$/;
  if (!passwordRegEx.test(value)) {
    callback('密码必须包含字母、数字,长度大于等于8个字符');
  } else {
    callback();
  }
};


const checkConfirmPassword = (_rule: any, value: string, callback: Function) => {
  if (value !== formState.newPassword) {
    callback('两次输入的密码不一致!');
  } else {
    callback();
  }
};
</script>

<style scoped>
.title-container {
  display: flex;
  flex-direction: column;
  justify-content: center;
  align-items: center;
  text-align: center;
  margin-top: 20px;
}

.input-field {
  /* margin-top: 20px; */
  height: 40px;
  width: 300px;
  border-radius: 10px;
  line-height: 40px !important;
}

.input-pass{
  height: 40px;
}

.submit-button {
  margin-top: 20px;
  height: 40px;
  width: 300px;
  border-radius: 10px;
}
</style>

async function encryptPassword(password: string): Promise<string> {
  try {
    // 获取后端公钥
    const response = await fetch("/api/public-key");
    if (!response.ok) {
      throw new Error(`HTTP error! status: ${response.status}`);
    }
    const data = await response.json();
    const publicKey = data.publicKey;

    // 验证公钥的有效性
    if (!publicKey || typeof publicKey !== 'string') {
      throw new Error("Invalid public key");
    }

    // 使用公钥加密密码
    const encryptor = new JSEncrypt();
    encryptor.setPublicKey(publicKey);
    const encrypted = encryptor.encrypt(password);
    if (!encrypted) {
      throw new Error("Failed to encrypt password");
    }
    return encrypted;
  } catch (error) {
    message.error("加密失败" + error);
    throw error;
  }
}