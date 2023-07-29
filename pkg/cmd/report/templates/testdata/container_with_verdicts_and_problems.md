
# <img height=20 src="https://listen.dev/assets/images/dolphin-noborder.png"> listen.dev ∙ Security Report
<table align=center>
  <tr>
    <td><b>critical</b> 🚨 6</td>
    <td><b>medium</b> ⚠️ 0</td>
    <td><b>low</b> 🔷 0</td>
  </tr>
</table>

### 🔍 The following behaviors have been detected in the dependency tree during installation
<details>
<summary>🚨 <b>Critical severity</b>
<table align="right">
<tr>
<td>📡📑</td>
<td>2 categories</td>
</tr>
</table>
</summary>
<br>

<ul>
  
<li>
<details>
<summary>
📡 <b>Dynamic instrumentation</b> ∙ 2 packages
</summary>
<br>

<ul>

<li>
<details>
<summary>📦 <i>foo@1.0.0</i> ∙ 5 occurrences ∙ 2 kind of issues ∙ <a href="https://verdicts.listen.dev/npm/foo/1.0.0">open 🔗</a>
</summary>
<br>    

<ul>

<li>
<details>
<summary>
<code>outbound network connection</code> ∙ 3 total occurrences
</summary>
<br>

| Name | Version | Transitive Dependency | Occurrences | More |
|---|---|---|---|---|
| foo | 1.0.0 || 3 | [🔗](https://verdicts.listen.dev/npm/foo/1.0.0) |

</details>
    
</li>

<li>
<details>
<summary>
<code>write to filesystem</code> ∙ 2 total occurrences
</summary>
<br>

| Name | Version | Transitive Dependency | Occurrences | More |
|---|---|---|---|---|
| bar | 1.0.0 |✔️| 1 | [🔗](https://verdicts.listen.dev/npm/bar/1.0.0) |
| foo | 1.0.0 || 1 | [🔗](https://verdicts.listen.dev/npm/foo/1.0.0) |

</details>
    
</li>

</ul>
</details>
</li>

<li>
<details>
<summary>📦 <i>baz@1.0.0</i> ∙ 1 occurrence ∙ 1 kind of issue ∙ <a href="https://verdicts.listen.dev/npm/baz/1.0.0">open 🔗</a>
</summary>
<br>    

<ul>

<li>
<details>
<summary>
<code>outbound network connection</code> ∙ 1 total occurrence
</summary>
<br>

| Name | Version | Transitive Dependency | Occurrences | More |
|---|---|---|---|---|
| baz | 1.0.0 || 1 | [🔗](https://verdicts.listen.dev/npm/baz/1.0.0) |

</details>
    
</li>

</ul>
</details>
</li>

</ul>
</details>    
</li>
  
<li>
<details>
<summary>
📑 <b>Metadata</b> ∙ 1 package
</summary>
<br>

<ul>

<li>
<details>
<summary>📦 <i>foo@1.0.0</i> ∙ 1 occurrence ∙ 1 kind of issue ∙ <a href="https://verdicts.listen.dev/npm/foo/1.0.0">open 🔗</a>
</summary>
<br>    

<ul>

<li>
<details>
<summary>
<code>missing description</code> ∙ 1 total occurrence
</summary>
<br>

| Name | Version | Transitive Dependency | Occurrences | More |
|---|---|---|---|---|
| bar | 1.0.0 |✔️| 1 | [🔗](https://verdicts.listen.dev/npm/bar/1.0.0) |

</details>
    
</li>

</ul>
</details>
</li>

</ul>
</details>    
</li>

</ul>
</details>
<hr>




### 🚩 Some problems have been encountered
<details>
<summary><a href="https://listen.dev/probs/invalid-name">🔗</a> <b>A problem that does not exist, just for testing</b> ∙ 2 occurrences ∙ <i>Package name not valid</i></summary>

- [foobar@1.0.0](https://verdicts.listen.dev/npm/foobar/1.0.0)
- [baz@1.0.0](https://verdicts.listen.dev/npm/baz/1.0.0)


[See docs 🔗](https://listen.dev/probs/invalid-name)
</details>
<details>
<summary><a href="https://listen.dev/probs/invalid-name">🔗</a> <b>Package name not valid</b> ∙ 2 occurrences ∙ <i>Package name not valid</i></summary>

- [foobar@1.0.0](https://verdicts.listen.dev/npm/foobar/1.0.0)
- [baz@1.0.0](https://verdicts.listen.dev/npm/baz/1.0.0)


[See docs 🔗](https://listen.dev/probs/invalid-name)
</details>

<hr>


<i>Powered by</i> <b><a href="https://listen.dev">listen.dev</a> <img height=14 src="https://listen.dev/assets/images/dolphin-noborder.png"></b>