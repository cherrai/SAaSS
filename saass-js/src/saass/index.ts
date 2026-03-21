import axios from 'axios'
import { ResponseData } from '@nyanyajs/utils/dist/nrequest'
import { NEventListener, Wait, NRequest } from '@nyanyajs/utils'

const CryptoJS = require('crypto-js')

export const getHash = (result: string | ArrayBuffer) => {
  const r: any = result
  const wordArray = CryptoJS.lib.WordArray.create(r)
  const hash = CryptoJS.SHA256(wordArray).toString()
  return hash
}
export type SType = 'Folder' | 'File'
export interface AccessToken {
  user: string
  temporaryAccessToken: string
}
export interface FolderItem {
  id: string
  folderName: string
  shortId: string
  parentFolderId: string
  path: string
  status: number
  availableRange: {
    authorId: string
    password: string
    allowShare: 1 | -1
    shareUsers: {
      uid: string
      createTime: number
    }[]
  }
  usage: {}
  createTime: number
  lastUpdateTime: number
  deleteTime: number
  accessToken?: AccessToken
}
export interface FileItem {
  id: string
  shortId: string
  fileName: string
  path: string
  availableRange: {
    visitCount: number
    expirationTime: number
    autoExtendPeriod: number
    authorId: string
    password: string
    allowShare: 1 | -1
    shareUsers: {
      uid: string
      createTime: number
    }[]
  }
  usage: {
    visitCount: number
  }
  createTime: number
  lastUpdateTime: number
  deleteTime: number
  urls: {
    domainUrl: string
    shortUrl: string
    url: string
  }
  fileInfo: {
    name: string
    size: number
    type: string
    suffix: string
    lastModified: number
    hash: string
    width: number
    height: number
  }
  accessToken?: AccessToken
}
export interface ChunkUploadTokenInfo {
  token: string
  uploadedOffset: number[]
  uploadedTotalSize: number
  chunkSize: number
  urls: {
    domainUrl: string
    uploadUrl: string
    shortUrl: string
    url: string
  }
}

export class SAaSS extends NEventListener<{
  AppTokenInvalid: SAaSS
}> {
  private baseUrl: string = ''
  private appToken: string = ''
  private appTokenDeadline: number = 0
  private wait = new Wait()
  private R = new NRequest()
  private timer?: NodeJS.Timeout

  constructor({ baseUrl }: { baseUrl?: string }) {
    super()
    baseUrl && (this.baseUrl = baseUrl)
  }
  private apiUrls = {
    v1: {
      versionPrefix: '/api/v1',
      createChunkupload: '/chunkupload/create',
      newFolder: '/folder/new',
      renameFolder: '/folder/rename',
      deleteFolders: '/folder/delete',
      moveFoldersToTrash: '/folder/moveToTrash',
      copyFolder: '/folder/copy',
      moveFolder: '/folder/move',
      restoreFolder: '/folder/restore',
      checkFolderExists: '/folder/checkExists',
      getFolderList: '/folder/list/get',
      getRecyclebinFolderList: '/folder/recyclebin/list/get',
      setFolderSharing: '/folder/share/set',
      setFolderPassword: '/folder/password/set',
      getFolderByShortId: '/folder/shortid/get',
      getFolderListWithShortId: '/folder/list/shortid/get',
      getRootFolderToken: '/folder/rootPathToken/get',

      getFileList: '/file/list/get',
      getFileByShortId: '/file/shortid/get',
      getRecentFiles: '/file/recent/list/get',
      getRecyclebinFiles: '/file/recyclebin/list/get',
      renameFile: '/file/rename',
      deleteFiles: '/file/delete',
      moveFilesToTrash: '/file/moveToTrash',
      restoreFile: '/file/restore',
      checkFileExists: '/file/checkExists',
      setFileSharing: '/file/share/set',
      setFileExpirationTime: '/file/expirationTime/set',
      setFileAutoExtendPeriod: '/file/autoExtendPeriod/set',
      setFilePassword: '/file/password/set',
      copyFile: '/file/copy',
      moveFile: '/file/move',
      getUserToken: '/app/userToken/get',
      getPasswordToken: '/file/passwordToken/get',
      getFileListWithShortId: '/file/list/shortid/get',
    },
  }
  setBaseUrl(baseUrl: string) {
    this.baseUrl = baseUrl
    this.wait.dispatch('baseUrl')
  }
  setAppToken(baseUrl: string, token: string, deadline: number) {
    // console.log('getRecentFiles', token)
    this.baseUrl = baseUrl
    this.appToken = token
    this.appTokenDeadline = deadline
    this.timer && clearTimeout(this.timer)
    this.timer = setTimeout(
      () => {
        this.wait.revoke()
        this.appToken = ''
        this.dispatch('AppTokenInvalid', this)
      },
      (deadline - 5) * 1000 - new Date().getTime()
    )
    this.wait.dispatch()
    this.wait.dispatch('baseUrl')
  }

  async getFileListWithShortId(
    id: string,
    accessToken: AccessToken,
    deadline: number
  ): Promise<FileItem[]> {
    // 未来支持联级创建
    await this.wait.waiting('baseUrl')
    const res = await this.R.request({
      method: 'GET',
      url:
        this.baseUrl +
        this.apiUrls.v1.versionPrefix +
        this.apiUrls.v1.getFileListWithShortId,
      data: {
        id,
        accessToken,
        deadline,
      },
      headers: {
        'Content-Type': 'multipart/form-data',
      },
    })
    if (res.data.code === 200) {
      return res.data?.data?.list.map((v: any): FileItem => {
        return {
          id: v.id,
          shortId: v.shortId,
          fileName: v.fileName,
          path: v.path,
          availableRange: v.availableRange,
          usage: v.usage,
          createTime: v.createTime,
          lastUpdateTime: v.lastUpdateTime,
          urls: v.urls,
          fileInfo: v.fileInfo,
          deleteTime: v.deleteTime,
          accessToken: v.accessToken,
        }
      })
    }
    return []
  }
  async getFileList(path: string): Promise<FileItem[]> {
    // 未来支持联级创建
    await this.wait.waiting()
    const res = await this.R.request({
      method: 'GET',
      url:
        this.baseUrl +
        this.apiUrls.v1.versionPrefix +
        this.apiUrls.v1.getFileList,
      data: {
        appToken: this.appToken,
        path,
      },
      headers: {
        'Content-Type': 'multipart/form-data',
      },
    })
    console.log(res.data)
    if (res.data.code === 200) {
      return res.data?.data?.list.map((v: any): FileItem => {
        return {
          id: v.id,
          shortId: v.shortId,
          fileName: v.fileName,
          path: v.path,
          availableRange: v.availableRange,
          usage: v.usage,
          createTime: v.createTime,
          lastUpdateTime: v.lastUpdateTime,
          urls: v.urls,
          fileInfo: v.fileInfo,
          deleteTime: v.deleteTime,
        }
      })
    }
    return []
  }
  async getFileByShortId(
    id: string,
    password: string = '',
    deadline: number
  ): Promise<ResponseData<FileItem | undefined>> {
    // 未来支持联级创建
    console.log('getFileByShortId', this.wait)
    await this.wait.waiting('baseUrl')
    const res = await this.R.request({
      method: 'GET',
      url:
        this.baseUrl +
        this.apiUrls.v1.versionPrefix +
        this.apiUrls.v1.getFileByShortId,
      data: {
        id,
        password,
        deadline,
      },
      headers: {
        'Content-Type': 'multipart/form-data',
      },
    })
    return res.data
  }
  async getFolderByShortId(
    id: string,
    password: string = '',
    deadline: number
  ): Promise<ResponseData<FolderItem | undefined>> {
    // 未来支持联级创建
    await this.wait.waiting('baseUrl')
    const res = await this.R.request({
      method: 'GET',
      url:
        this.baseUrl +
        this.apiUrls.v1.versionPrefix +
        this.apiUrls.v1.getFolderByShortId,
      data: {
        id,
        password,
        deadline,
      },
      headers: {
        'Content-Type': 'multipart/form-data',
      },
    })
    return res.data
  }
  async getFolderListWithShortId(
    id: string,
    accessToken: AccessToken,
    deadline: number
  ): Promise<FolderItem[]> {
    // 未来支持联级创建
    await this.wait.waiting('baseUrl')
    const res = await this.R.request({
      method: 'GET',
      url:
        this.baseUrl +
        this.apiUrls.v1.versionPrefix +
        this.apiUrls.v1.getFolderListWithShortId,
      data: {
        id,
        accessToken,
        deadline,
      },
      headers: {
        'Content-Type': 'multipart/form-data',
      },
    })
    if (res.data.code === 200) {
      return (
        res.data?.data?.list?.map((v: any): FolderItem => {
          return {
            id: v.id,
            shortId: v.shortId,
            folderName: v.folderName,
            path: v.path,
            parentFolderId: v.parentFolderId,
            availableRange: v.availableRange,
            usage: v.usage,
            status: v.status,
            createTime: v.createTime,
            lastUpdateTime: v.lastUpdateTime,
            deleteTime: v.deleteTime,
            accessToken: v.accessToken,
          }
        }) || []
      )
    }
    return []
  }
  async getRecentFiles(
    path: string,
    pageNum: number,
    pageSize: number,
    keywords: string[] = []
  ): Promise<FileItem[]> {
    // 未来支持联级创建
    await this.wait.waiting()
    console.log({
      appToken: this.appToken,
      path,
      pageNum,
      pageSize,
      keywords: keywords?.join(','),
    })
    const res = await this.R.request({
      method: 'GET',
      url:
        this.baseUrl +
        this.apiUrls.v1.versionPrefix +
        this.apiUrls.v1.getRecentFiles,
      data: {
        appToken: this.appToken,
        path,
        pageNum,
        pageSize,
        keywords: keywords?.join(','),
      },
      headers: {
        'Content-Type': 'multipart/form-data',
      },
    })
    if (res.data.code === 200) {
      return res.data?.data?.list.map((v: any): FileItem => {
        return {
          id: v.id,
          shortId: v.shortId,
          fileName: v.fileName,
          path: v.path,
          availableRange: v.availableRange,
          usage: v.usage,
          createTime: v.createTime,
          lastUpdateTime: v.lastUpdateTime,
          urls: v.urls,
          fileInfo: v.fileInfo,
          deleteTime: v.deleteTime,
        }
      })
    }
    return []
  }
  async getRecyclebinFiles(
    path: string,
    pageNum: number,
    pageSize: number
  ): Promise<FileItem[]> {
    // 未来支持联级创建
    await this.wait.waiting()
    const res = await this.R.request({
      method: 'GET',
      url:
        this.baseUrl +
        this.apiUrls.v1.versionPrefix +
        this.apiUrls.v1.getRecyclebinFiles,
      data: {
        appToken: this.appToken,
        path,
        pageNum,
        pageSize,
      },
      headers: {
        'Content-Type': 'multipart/form-data',
      },
    })
    if (res.data.code === 200) {
      if (!res.data?.data?.total) return []
      return res.data?.data?.list.map((v: any): FileItem => {
        return {
          id: v.id,
          shortId: v.shortId,
          fileName: v.fileName,
          path: v.path,
          availableRange: v.availableRange,
          usage: v.usage,
          createTime: v.createTime,
          lastUpdateTime: v.lastUpdateTime,
          urls: v.urls,
          fileInfo: v.fileInfo,
          deleteTime: v.deleteTime,
        }
      })
    } else {
      console.error(res.data)
    }
    return []
  }
  async getRecyclebinFolderList(
    path: string,
    pageNum: number,
    pageSize: number
  ): Promise<FolderItem[]> {
    // 未来支持联级创建
    await this.wait.waiting()
    const res = await this.R.request({
      method: 'GET',
      url:
        this.baseUrl +
        this.apiUrls.v1.versionPrefix +
        this.apiUrls.v1.getRecyclebinFolderList,
      data: {
        appToken: this.appToken,
        path,
        pageNum,
        pageSize,
      },
      headers: {
        'Content-Type': 'multipart/form-data',
      },
    })
    if (res.data.code === 200) {
      if (!res.data?.data?.total) return []
      return (
        res.data?.data?.list?.map((v: any): FolderItem => {
          return {
            id: v.id,
            shortId: v.shortId,
            folderName: v.folderName,
            path: v.path,
            parentFolderId: v.parentFolderId,
            availableRange: v.availableRange,
            usage: v.usage,
            status: v.status,
            createTime: v.createTime,
            lastUpdateTime: v.lastUpdateTime,
            deleteTime: v.deleteTime,
          }
        }) || []
      )
    } else {
      console.error(res.data)
    }
    return []
  }
  async getFolderList(parentPath: string): Promise<FolderItem[]> {
    // 未来支持联级创建
    await this.wait.waiting()
    const res = await this.R.request({
      method: 'GET',
      url:
        this.baseUrl +
        this.apiUrls.v1.versionPrefix +
        this.apiUrls.v1.getFolderList,
      data: {
        appToken: this.appToken,
        parentPath,
      },
      headers: {
        'Content-Type': 'multipart/form-data',
      },
    })
    if (res.data.code === 200) {
      if (!res.data?.data?.total) return []
      return (
        res.data?.data?.list?.map((v: any): FolderItem => {
          return {
            id: v.id,
            shortId: v.shortId,
            folderName: v.folderName,
            path: v.path,
            parentFolderId: v.parentFolderId,
            availableRange: v.availableRange,
            usage: v.usage,
            status: v.status,
            createTime: v.createTime,
            lastUpdateTime: v.lastUpdateTime,
            deleteTime: v.deleteTime,
          }
        }) || []
      )
    }
    return []
  }
  async newFolder(folderName: string, parentPath: string) {
    // 未来支持联级创建
    await this.wait.waiting()

    return (
      await this.R.request({
        method: 'POST',
        url:
          this.baseUrl +
          this.apiUrls.v1.versionPrefix +
          this.apiUrls.v1.newFolder,
        data: {
          appToken: this.appToken,
          parentPath,
          folderName,
        },
        headers: {
          'Content-Type': 'multipart/form-data',
        },
      })
    ).data
  }
  async renameFile(path: string, oldFileName: string, newFileName: string) {
    // 未来支持联级创建
    await this.wait.waiting()

    return (
      await this.R.request({
        method: 'POST',
        url:
          this.baseUrl +
          this.apiUrls.v1.versionPrefix +
          this.apiUrls.v1.renameFile,
        data: {
          appToken: this.appToken,
          path,
          oldFileName,
          newFileName,
        },
        headers: {
          'Content-Type': 'multipart/form-data',
        },
      })
    ).data
  }
  async moveFilesToTrash(path: string, fileNames: string[]) {
    // 未来支持联级创建
    await this.wait.waiting()

    return (
      await this.R.request({
        method: 'POST',
        url:
          this.baseUrl +
          this.apiUrls.v1.versionPrefix +
          this.apiUrls.v1.moveFilesToTrash,
        data: {
          appToken: this.appToken,
          path: path,
          fileNames: fileNames,
        },
        headers: {
          'Content-Type': 'multipart/form-data',
        },
      })
    ).data
  }
  async checkFileExists(path: string, fileNames: string[]) {
    // 未来支持联级创建
    await this.wait.waiting()

    const res = await this.R.request({
      method: 'POST',
      url:
        this.baseUrl +
        this.apiUrls.v1.versionPrefix +
        this.apiUrls.v1.checkFileExists,
      data: {
        appToken: this.appToken,
        path: path,
        fileNames: fileNames,
      },
      headers: {
        'Content-Type': 'multipart/form-data',
      },
    })
    if (res.data.code === 200) {
      if (!res.data?.data?.total) return []
      return res.data?.data?.list as string[]
    }
    return []
  }
  async restoreFile(path: string, fileNames: string[]) {
    // 未来支持联级创建
    await this.wait.waiting()

    return (
      await this.R.request({
        method: 'POST',
        url:
          this.baseUrl +
          this.apiUrls.v1.versionPrefix +
          this.apiUrls.v1.restoreFile,
        data: {
          appToken: this.appToken,
          path: path,
          fileNames: fileNames,
        },
        headers: {
          'Content-Type': 'multipart/form-data',
        },
      })
    ).data
  }
  async deleteFiles(path: string, fileNames: string[]) {
    // 未来支持联级创建
    await this.wait.waiting()

    return (
      await this.R.request({
        method: 'POST',
        url:
          this.baseUrl +
          this.apiUrls.v1.versionPrefix +
          this.apiUrls.v1.deleteFiles,
        data: {
          appToken: this.appToken,
          path: path,
          fileNames: fileNames,
        },
        headers: {
          'Content-Type': 'multipart/form-data',
        },
      })
    ).data
  }

  async setFileSharing(path: string, fileNames: string[], status: 1 | -1) {
    // 未来支持联级创建
    await this.wait.waiting()

    return (
      await this.R.request({
        method: 'POST',
        url:
          this.baseUrl +
          this.apiUrls.v1.versionPrefix +
          this.apiUrls.v1.setFileSharing,
        data: {
          appToken: this.appToken,
          path,
          fileNames,
          status,
        },
        headers: {
          'Content-Type': 'multipart/form-data',
        },
      })
    ).data
  }

  async setFileExpirationTime(
    path: string,
    fileNames: string[],
    expirationTime: number = -1
  ) {
    // 未来支持联级创建
    await this.wait.waiting()

    return (
      await this.R.request({
        method: 'POST',
        url:
          this.baseUrl +
          this.apiUrls.v1.versionPrefix +
          this.apiUrls.v1.setFileExpirationTime,
        data: {
          appToken: this.appToken,
          path,
          fileNames,
          expirationTime,
        },
        headers: {
          'Content-Type': 'multipart/form-data',
        },
      })
    ).data
  }

  async setFileAutoExtendPeriod(
    path: string,
    fileNames: string[],
    autoExtendPeriod: number = -1
  ) {
    // 未来支持联级创建
    await this.wait.waiting()

    return (
      await this.R.request({
        method: 'POST',
        url:
          this.baseUrl +
          this.apiUrls.v1.versionPrefix +
          this.apiUrls.v1.setFileAutoExtendPeriod,
        data: {
          appToken: this.appToken,
          path,
          fileNames,
          autoExtendPeriod,
        },
        headers: {
          'Content-Type': 'multipart/form-data',
        },
      })
    ).data
  }

  async setFilePassword(path: string, fileName: string, password: string) {
    // 未来支持联级创建
    await this.wait.waiting()

    return (
      await this.R.request({
        method: 'POST',
        url:
          this.baseUrl +
          this.apiUrls.v1.versionPrefix +
          this.apiUrls.v1.setFilePassword,
        data: {
          appToken: this.appToken,
          path,
          fileName,
          password,
        },
        headers: {
          'Content-Type': 'multipart/form-data',
        },
      })
    ).data
  }

  async setFolderSharing(path: string, folderNames: string[], status: 1 | -1) {
    // 未来支持联级创建
    await this.wait.waiting()

    return (
      await this.R.request({
        method: 'POST',
        url:
          this.baseUrl +
          this.apiUrls.v1.versionPrefix +
          this.apiUrls.v1.setFolderSharing,
        data: {
          appToken: this.appToken,
          path,
          folderNames,
          status,
        },
        headers: {
          'Content-Type': 'multipart/form-data',
        },
      })
    ).data
  }
  async setFolderPassword(path: string, folderName: string, password: string) {
    // 未来支持联级创建
    await this.wait.waiting()

    return (
      await this.R.request({
        method: 'POST',
        url:
          this.baseUrl +
          this.apiUrls.v1.versionPrefix +
          this.apiUrls.v1.setFolderPassword,
        data: {
          appToken: this.appToken,
          path,
          folderName,
          password,
        },
        headers: {
          'Content-Type': 'multipart/form-data',
        },
      })
    ).data
  }
  async copyFile(path: string, fileNames: string[], newPath: string) {
    // 未来支持联级创建
    await this.wait.waiting()

    return (
      await this.R.request({
        method: 'POST',
        url:
          this.baseUrl +
          this.apiUrls.v1.versionPrefix +
          this.apiUrls.v1.copyFile,
        data: {
          appToken: this.appToken,
          path,
          fileNames,
          newPath,
        },
        headers: {
          'Content-Type': 'multipart/form-data',
        },
      })
    ).data
  }

  async moveFile(path: string, fileNames: string[], newPath: string) {
    // 未来支持联级创建
    await this.wait.waiting()

    return (
      await this.R.request({
        method: 'POST',
        url:
          this.baseUrl +
          this.apiUrls.v1.versionPrefix +
          this.apiUrls.v1.moveFile,
        data: {
          appToken: this.appToken,
          path,
          fileNames,
          newPath,
        },
        headers: {
          'Content-Type': 'multipart/form-data',
        },
      })
    ).data
  }

  async renameFolder(
    parentPath: string,
    oldFolderName: string,
    newFolderName: string
  ) {
    // 未来支持联级创建
    await this.wait.waiting()

    return (
      await this.R.request({
        method: 'POST',
        url:
          this.baseUrl +
          this.apiUrls.v1.versionPrefix +
          this.apiUrls.v1.renameFolder,
        data: {
          appToken: this.appToken,
          parentPath,
          oldFolderName,
          newFolderName,
        },
        headers: {
          'Content-Type': 'multipart/form-data',
        },
      })
    ).data
  }
  async moveFoldersToTrash(parentPath: string, folderNames: string[]) {
    // 未来支持联级创建
    await this.wait.waiting()

    return (
      await this.R.request({
        method: 'POST',
        url:
          this.baseUrl +
          this.apiUrls.v1.versionPrefix +
          this.apiUrls.v1.moveFoldersToTrash,
        data: {
          appToken: this.appToken,
          parentPath,
          folderNames,
        },
        headers: {
          'Content-Type': 'multipart/form-data',
        },
      })
    ).data
  }
  async checkFolderExists(parentPath: string, folderNames: string[]) {
    // 未来支持联级创建
    await this.wait.waiting()
    const res = await this.R.request({
      method: 'POST',
      url:
        this.baseUrl +
        this.apiUrls.v1.versionPrefix +
        this.apiUrls.v1.checkFolderExists,
      data: {
        appToken: this.appToken,
        parentPath,
        folderNames,
      },
      headers: {
        'Content-Type': 'multipart/form-data',
      },
    })
    if (res.data.code === 200) {
      if (!res.data?.data?.total) return []
      return res.data?.data?.list as string[]
    }
    return []
  }
  async restoreFolder(parentPath: string, folderNames: string[]) {
    // 未来支持联级创建
    await this.wait.waiting()

    return (
      await this.R.request({
        method: 'POST',
        url:
          this.baseUrl +
          this.apiUrls.v1.versionPrefix +
          this.apiUrls.v1.restoreFolder,
        data: {
          appToken: this.appToken,
          parentPath,
          folderNames,
        },
        headers: {
          'Content-Type': 'multipart/form-data',
        },
      })
    ).data
  }
  async deleteFolders(parentPath: string, folderNames: string[]) {
    // 未来支持联级创建
    await this.wait.waiting()

    return (
      await this.R.request({
        method: 'POST',
        url:
          this.baseUrl +
          this.apiUrls.v1.versionPrefix +
          this.apiUrls.v1.deleteFolders,
        data: {
          appToken: this.appToken,
          parentPath,
          folderNames,
        },
        headers: {
          'Content-Type': 'multipart/form-data',
        },
      })
    ).data
  }
  async copyFolder(
    parentPath: string,
    folderNames: string[],
    newParentPath: string
  ) {
    // 未来支持联级创建
    await this.wait.waiting()

    return (
      await this.R.request({
        method: 'POST',
        url:
          this.baseUrl +
          this.apiUrls.v1.versionPrefix +
          this.apiUrls.v1.copyFolder,
        data: {
          appToken: this.appToken,
          parentPath,
          folderNames,
          newParentPath,
        },
        headers: {
          'Content-Type': 'multipart/form-data',
        },
      })
    ).data
  }

  async moveFolder(
    parentPath: string,
    folderNames: string[],
    newParentPath: string
  ) {
    // 未来支持联级创建
    await this.wait.waiting()

    return (
      await this.R.request({
        method: 'POST',
        url:
          this.baseUrl +
          this.apiUrls.v1.versionPrefix +
          this.apiUrls.v1.moveFolder,
        data: {
          appToken: this.appToken,
          parentPath,
          folderNames,
          newParentPath,
        },
        headers: {
          'Content-Type': 'multipart/form-data',
        },
      })
    ).data
  }
  async getUserToken(): Promise<string> {
    // 未来支持联级创建
    await this.wait.waiting()

    const res = (
      await this.R.request({
        method: 'POST',
        url:
          this.baseUrl +
          this.apiUrls.v1.versionPrefix +
          this.apiUrls.v1.getUserToken,
        data: {
          appToken: this.appToken,
        },
        headers: {
          'Content-Type': 'multipart/form-data',
        },
      })
    ).data
    if (res.code === 200) {
      return res.data.token
    } else {
      throw res
    }
  }
  async getPasswordToken(
    path: string,
    fileName: string,
    deadline: number
  ): Promise<{
    user: string
    passwordToken: string
    // rootPathToken: string
  }> {
    // 未来支持联级创建
    await this.wait.waiting()

    const res = (
      await this.R.request({
        method: 'POST',
        url:
          this.baseUrl +
          this.apiUrls.v1.versionPrefix +
          this.apiUrls.v1.getPasswordToken,
        data: {
          appToken: this.appToken,
          path,
          fileName,
          deadline,
        },
        headers: {
          'Content-Type': 'multipart/form-data',
        },
      })
    ).data
    if (res.code === 200) {
      return res.data
    } else {
      throw res
    }
  }
  // async getRootFolderToken(deadline: number): Promise<{
  // 	rootPathToken: string
  // }> {
  // 	// 未来支持联级创建
  // 	await this.wait.waiting()

  // 	const res = (
  // 		await this.R.request({
  // 			method: 'POST',
  // 			url:
  // 				this.baseUrl +
  // 				this.apiUrls.v1.versionPrefix +
  // 				this.apiUrls.v1.getRootFolderToken,
  // 			data: {
  // 				appToken: this.appToken,
  // 				deadline,
  // 			},
  // 			headers: {
  // 				'Content-Type': 'multipart/form-data',
  // 			},
  // 		})
  // 	).data
  // 	if (res.code === 200) {
  // 		return res.data
  // 	} else {
  // 		throw res
  // 	}
  // }

  // 未来改成可以同时上传多个文件，但这样断电断续可能会有点问题，
  // 那就改成，可以10MB或者50MB的速度上传，未来再说
  uploadFile(options: {
    file: File
    url: string
    token: string
    chunkSize: number
    uploadedOffset: number[]
    uploadedTotalSize: number
    onprogress: (options: { uploadedSize: number; totalSize: number }) => void
    onsuccess: (options: { shortUrl: string; url: string }) => void
    onerror: (err: string) => void
  }) {
    // console.log('options', options)

    let offset = options.uploadedTotalSize
    let chunkSize = options.chunkSize

    const reader = new FileReader()
    reader.onload = async (e) => {
      if (!e.target?.result) return
      // console.log(options, offset, offset + chunkSize)
      const result: any = e.target?.result
      const hash = CryptoJS.SHA256(
        CryptoJS.lib.WordArray.create(result)
      ).toString()
      const blob = new File(
        [e.target.result],
        encodeURIComponent(
          JSON.stringify({
            offset: offset.toString(),
            hash: hash,
            // 有问题
            final: e.total + offset === options.file.size ? 'ok' : 'no',
          })
        )
      )
      const res = await api.uploadFile(
        options.url,
        options.token,
        blob,
        (progress) => {
          options.onprogress({
            uploadedSize: offset + e.total * (progress / 100),
            totalSize: options.file.size,
          })
        }
      )
      // console.log(res)
      if (res.data.code === 200) {
        // options.onprogress({
        // 	uploadedSize: e.total + offset,
        // 	totalSize: options.file.size,
        // })
        // console.log(e.total + offset, options.file.size)
        if (e.total + offset === options.file.size) {
          options.onsuccess({
            shortUrl: res.data.data.shortUrl,
            url: res.data.data.url,
          })
          return
        }
        offset = offset + chunkSize
        reader.readAsArrayBuffer(options.file.slice(offset, offset + chunkSize))
      } else {
        options.onerror(res.data.error || res.data.msg)
      }
    }

    reader.readAsArrayBuffer(options.file.slice(offset, offset + chunkSize))
  }

  getHash(result: string | ArrayBuffer) {
    const r: any = result
    const wordArray = CryptoJS.lib.WordArray.create(r)
    const hash = CryptoJS.SHA256(wordArray).toString()
    return hash
  }
  async createChunkUpload({
    folderPath,
    fileName,
    chunkSize,
    visitCount = -1,
    password = '',
    expirationTime = -1,
    autoExtendPeriod = -1,
    fileInfo,
    fileConflict,
    allowShare,
    shareUsers = ['AllUser'],
  }: {
    // 在服务端需要自动加上rootPath,参考folder
    folderPath: string
    fileName: string
    // 16 * 1024,
    // 32 * 1024,
    // 64 * 1024,
    // 128 * 1024,
    // 256 * 1024,
    // 512 * 1024,
    // 1024 * 1024,
    // 2 * 1024 * 1024,
    // 5 * 1024 * 1024,
    // 10 * 1024 * 1024,
    // 20 * 1024 * 1024,
    // 30 * 1024 * 1024,
    // 50 * 1024 * 1024
    chunkSize?: number
    visitCount?: number
    password?: string
    // -1
    expirationTime?: number
    autoExtendPeriod?: number
    fileInfo: {
      name: string
      size: number
      type: string
      fileSuffix: string
      lastModified: number
      hash: string
      exif?: string
    }
    fileConflict: 'Replace' | 'Skip'
    allowShare: 1 | -1
    shareUsers?: string[]
  }): Promise<ChunkUploadTokenInfo | undefined> {
    // 未来支持联级创建
    await this.wait.waiting()

    if (!chunkSize) {
      chunkSize = 128 * 1024

      // console.log(fileInfo.size, fileInfo.size / 1024 / 1024)
      // 大于10MB
      if (fileInfo.size > 10 * 1024 * 1024) {
        chunkSize = 512 * 1024
      }

      // 大于30MB
      if (fileInfo.size > 30 * 1024 * 1024) {
        chunkSize = 1 * 1024 * 1024
      }
      if (fileInfo.size > 50 * 1024 * 1024) {
        chunkSize = 2 * 1024 * 1024
      }
      if (fileInfo.size > 100 * 1024 * 1024) {
        chunkSize = 5 * 1024 * 1024
      }
      if (fileInfo.size > 200 * 1024 * 1024) {
        chunkSize = 10 * 1024 * 1024
      }
    }
    const res = (
      await this.R.request({
        method: 'POST',
        url:
          this.baseUrl +
          this.apiUrls.v1.versionPrefix +
          this.apiUrls.v1.createChunkupload,
        data: {
          appToken: this.appToken,
          path: folderPath,
          fileName: fileName,
          chunkSize: chunkSize,
          visitCount: visitCount,
          password: password,
          expirationTime: expirationTime,
          autoExtendPeriod: autoExtendPeriod,
          // "type":           options.Type,
          fileInfo: fileInfo,
          fileConflict: fileConflict,
          allowShare,
          shareUsers,
        },
        headers: {
          'Content-Type': 'multipart/form-data',
        },
      })
    ).data
    if (res.code === 200) {
      return {
        ...res.data,
        chunkSize: chunkSize,
        urls: {
          ...res.data.urls,
          uploadUrl: res.data.urls.domainUrl + '/api/v1/chunkupload/upload',
        },
      }
    }
    throw res
  }
  getType(id: string): SType {
    const t = id.substring(0, 1)
    if (['A', 'B', 'C', 'D', 'E', 'F', 'G'].includes(t)) {
      return 'Folder'
    } else {
      return 'File'
    }
  }
}

export const api = {
  async uploadFile(
    apiUrl: string,
    token: string,
    blob: File,
    onUploadProgress: (progress: number) => void
  ) {
    let formData = new FormData()
    formData.append('files', blob)
    return await axios.post(apiUrl + '?token=' + token, formData, {
      headers: {
        'Content-Type': 'multipart/form-data',
      },
      onUploadProgress: (e) => {
        if (e.lengthComputable) {
          onUploadProgress(((e.loaded / e.total) * 100) | 0)
        }
      },
    })
  },
}

// api.uploadFile(options.url, options.token)
// options.onprogress({
// 	uploadedSize: 3000,
// 	totalSize: options.file.size,
// })
// options.onsuccess({
// 	shortUrl: '',
// 	url: '',
// })
// options.onerror()
export default {
  getHash,
  api,
  SAaSS,
}
