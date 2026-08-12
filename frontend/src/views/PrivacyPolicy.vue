<template>
  <iframe
    :src="privacyPolicyUrl"
    style="width: 100%; height: 100%; position: absolute; left: 0"
    frameborder="0"
  ></iframe>
</template>

<script>
import { getPrivacyPolicyUrl } from "@/utils/runtime_config"

export default {
  name: "PrivacyPolicy",

  metaInfo: {
    title: "Privacy Policy - Timeful",
  },

  computed: {
    privacyPolicyUrl() {
      return getPrivacyPolicyUrl()
    },
  },
}
</script>
