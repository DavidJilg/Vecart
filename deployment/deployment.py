import errno
import os
import shutil
import stat
import subprocess

from PIL import Image


def handleRemoveReadonly(func, path, exc) -> None:
    excvalue = exc[1]
    if func in (os.rmdir, os.remove) and excvalue.errno == errno.EACCES:
        os.chmod(path, stat.S_IRWXU | stat.S_IRWXG | stat.S_IRWXO)  # 0777
        func(path)
    else:
        raise


def generate_windows_icon(deployment_dir: str) -> None:
    tmp_dir = deployment_dir + "tmp/"

    img = Image.open(deployment_dir + "../images/logo.png")

    if not os.path.isdir(tmp_dir):
        os.mkdir(tmp_dir)
    else:
        shutil.rmtree(tmp_dir, ignore_errors=False, onerror=handleRemoveReadonly)
        os.mkdir(tmp_dir)

    img_resized = img.resize(size=(256, 256), resample=Image.Resampling.LANCZOS)
    img_resized.save(fp=tmp_dir + f"icon.ico")


def get_version(deployment_dir: str) -> str:
    with open(f"{deployment_dir}../main.go", "r") as main_file:
        content = main_file.readlines()
        for line in content:
            if "const Version" in line:
                return f"v{line.split('"')[-2]}"
    return ""


def build_vecart(deployment_dir: str, version: str, deployment_targets: [(str, str, str)]) -> None:
    tmp_dir = deployment_dir + "tmp/"
    main_dir = deployment_dir + "../"

    for deployment_target in deployment_targets:
        os.environ["GOOS"] = deployment_target[0]
        os.environ["GOARCH"] = deployment_target[1]
        subprocess.run(["go", "build"], cwd=main_dir, env=os.environ)
        if deployment_target[0] == "windows":
            shutil.copyfile(main_dir + "Vecart.exe",
                            f"{tmp_dir}Vecart_{version}_{deployment_target[2]}.exe")
            os.remove(main_dir + "Vecart.exe")
            add_icon_to_exe(deployment_dir, f"{tmp_dir}Vecart_{version}_{deployment_target[2]}.exe")
        else:
            shutil.copyfile(main_dir + "Vecart",
                            f"{tmp_dir}Vecart_{version}_{deployment_target[2]}")
            os.remove(main_dir + "Vecart")


def add_icon_to_exe(deployment_dir: str, exe: str) -> None:
    subprocess.run(
        args=[f"ResourceHacker.exe",
              "-open", exe,
              "-save", exe,
              "-action", "addskip",
              "-res", "tmp/icon.ico",
              "-mask", "ICONGROUP,MAINICON"],
        cwd=deployment_dir)


if __name__ == '__main__':
    deployment_dir = os.path.dirname(os.path.realpath(__file__)) + "/"
    version = get_version(deployment_dir)
    deployment_targets = [
        ("windows", "386", "Windows_x86"),
        ("windows", "amd64", "Windows_x86-64"),
        ("linux", "386", "Linux_x86"),
        ("linux", "amd64", "Linux_x86-64")
    ]

    generate_windows_icon(deployment_dir)
    build_vecart(deployment_dir, version, deployment_targets)
