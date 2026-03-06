// 文件上传示例代码

// 1. 上传文件
async function uploadFile(file, token) {
  const formData = new FormData();
  formData.append('file', file);

  const response = await fetch('http://localhost:8080/_api/v1/files', {
    method: 'POST',
    headers: {
      'Authorization': `Bearer ${token}`
    },
    body: formData
  });

  if (!response.ok) {
    throw new Error('File upload failed');
  }

  return await response.json();
}

// 2. 发送文件消息（私聊）
async function sendFileMessage(recipient, fileInfo, token) {
  const response = await fetch('http://localhost:8080/_api/v1/messages', {
    method: 'POST',
    headers: {
      'Authorization': `Bearer ${token}`,
      'Content-Type': 'application/json'
    },
    body: JSON.stringify({
      recipient: recipient,
      content: {
        msgtype: 'm.file',
        body: fileInfo.filename,
        url: fileInfo.url,
        info: {
          size: fileInfo.file_size,
          mimetype: fileInfo.content_type
        }
      }
    })
  });

  if (!response.ok) {
    throw new Error('Failed to send message');
  }

  return await response.json();
}

// 3. 发送文件消息（群聊）
async function sendRoomFileMessage(roomId, fileInfo, token) {
  const response = await fetch(`http://localhost:8080/_api/v1/rooms/${roomId}/messages`, {
    method: 'POST',
    headers: {
      'Authorization': `Bearer ${token}`,
      'Content-Type': 'application/json'
    },
    body: JSON.stringify({
      content: {
        msgtype: 'm.file',
        body: fileInfo.filename,
        url: fileInfo.url,
        info: {
          size: fileInfo.file_size,
          mimetype: fileInfo.content_type
        }
      }
    })
  });

  if (!response.ok) {
    throw new Error('Failed to send room message');
  }

  return await response.json();
}

// 4. 下载文件
async function downloadFile(fileId, token) {
  const response = await fetch(`http://localhost:8080/_api/v1/files/${fileId}`, {
    method: 'GET',
    headers: {
      'Authorization': `Bearer ${token}`
    }
  });

  if (!response.ok) {
    throw new Error('File download failed');
  }

  // 获取文件名
  const contentDisposition = response.headers.get('Content-Disposition');
  const filenameMatch = contentDisposition?.match(/filename="(.+)"/);
  const filename = filenameMatch ? filenameMatch[1] : 'download';

  // 创建下载链接
  const blob = await response.blob();
  const url = window.URL.createObjectURL(blob);
  const a = document.createElement('a');
  a.href = url;
  a.download = filename;
  document.body.appendChild(a);
  a.click();
  window.URL.revokeObjectURL(url);
  document.body.removeChild(a);
}

// 5. 完整流程示例：上传并发送文件
async function uploadAndSendFile(file, recipient, token) {
  try {
    // 上传文件
    console.log('Uploading file...');
    const fileInfo = await uploadFile(file, token);
    console.log('File uploaded:', fileInfo);

    // 发送消息
    console.log('Sending message...');
    const message = await sendFileMessage(recipient, fileInfo, token);
    console.log('Message sent:', message);

    return { fileInfo, message };
  } catch (error) {
    console.error('Error:', error);
    throw error;
  }
}

// 6. HTML 表单示例
/*
<form id="fileUploadForm">
  <input type="file" id="fileInput" required>
  <input type="text" id="recipient" placeholder="@user:domain" required>
  <button type="submit">Send File</button>
</form>

<script>
document.getElementById('fileUploadForm').addEventListener('submit', async (e) => {
  e.preventDefault();

  const file = document.getElementById('fileInput').files[0];
  const recipient = document.getElementById('recipient').value;
  const token = localStorage.getItem('access_token'); // 从登录获取

  try {
    await uploadAndSendFile(file, recipient, token);
    alert('File sent successfully!');
  } catch (error) {
    alert('Failed to send file: ' + error.message);
  }
});
</script>
*/

// 7. React 组件示例
/*
import React, { useState } from 'react';

function FileUpload({ recipient, token }) {
  const [file, setFile] = useState(null);
  const [uploading, setUploading] = useState(false);

  const handleFileChange = (e) => {
    setFile(e.target.files[0]);
  };

  const handleSubmit = async (e) => {
    e.preventDefault();
    if (!file) return;

    setUploading(true);
    try {
      await uploadAndSendFile(file, recipient, token);
      alert('File sent successfully!');
      setFile(null);
    } catch (error) {
      alert('Failed to send file: ' + error.message);
    } finally {
      setUploading(false);
    }
  };

  return (
    <form onSubmit={handleSubmit}>
      <input
        type="file"
        onChange={handleFileChange}
        disabled={uploading}
      />
      <button type="submit" disabled={!file || uploading}>
        {uploading ? 'Sending...' : 'Send File'}
      </button>
    </form>
  );
}
*/

export {
  uploadFile,
  sendFileMessage,
  sendRoomFileMessage,
  downloadFile,
  uploadAndSendFile
};
