<script>
    export let value;
  
</script>

<select class="border w-full" bind:value={value}>
    <option value="SOCKS4">SOCKS4</option>
    <option value="SOCKS5">SOCKS5</option>
    <option value="SSH(Password)">SSH(Password)</option>
    <option value="SSH(PublicKeys)">SSH(PublicKeys)</option>
    <option value="UNENCRYPTED">ss(UNENCRYPTED)</option>
    <option value="AES-128-CTR">ss(AES-128-CTR)</option>
    <option value="AES-192-CTR">ss(AES-192-CTR)</option>
    <option value="AES-256-CTR">ss(AES-256-CTR)</option>
    <option value="AES-128-CFB">ss(AES-128-CFB)</option>
    <option value="AES-192-CFB">ss(AES-192-CFB)</option>
    <option value="AES-256-CFB">ss(AES-256-CFB)</option>
    <option value="XCHACHA20">ss(XCHACHA20)</option>
    <option value="XCHACHA20-IETF">ss(XCHACHA20-IETF)</option>
</select>

