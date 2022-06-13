<script>
    import Layout from "./lib/layout.svelte";
</script>

<Layout>
    <div>
        <p>代理命中流程</p>
        <p>1. 通过Rules命中域名</p>
        <p>2.1 如果命中，进行远端DNS解析(如果有配置)</p>
        <p>2.2 远端DNS也将重复代理流得程</p>
        <p>2.3 通过代理访部</p>
        <p>3.1 如果不命中，进行本地DNS解析(如果有配置)</p>
        <p>3.2 对解析到的IP重复代理流程</p>
        <p>4. 其它情况不走代理</p>
    </div>
    <br />
    <br />
    <div>
        <p>Rules</p>
        <p>1. 完整域名匹配</p>
        <p>www.google.com</p>
        <p>2. 正则匹配</p>
        <p>%google\.com$%</p>
        <p>3. IP段匹配</p>
        <p>127.0.0.1/8</p>
        <p>192.168.0.1/24</p>
    </div>
</Layout>

<style>
</style>
