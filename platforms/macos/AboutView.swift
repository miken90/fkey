import SwiftUI

struct AboutView: View {
    var body: some View {
        VStack(spacing: 0) {
            content
            Divider()
            footer
        }
        .frame(width: 320)
    }

    private var content: some View {
        VStack(spacing: 12) {
            Spacer()

            // App info
            Image(nsImage: AppMetadata.logo)
                .resizable()
                .frame(width: 72, height: 72)
            Text(AppMetadata.name)
                .font(.system(size: 20, weight: .bold))
            Text("Gõ Nhanh")
                .font(.caption)
                .foregroundStyle(.secondary)
            Text("Version \(AppMetadata.version)")
                .font(.caption)
                .foregroundStyle(.tertiary)

            Spacer()

            // Author
            VStack(spacing: 4) {
                Text(AppMetadata.author).font(.callout).fontWeight(.medium)
                Link(AppMetadata.authorEmail, destination: URL(string: "mailto:\(AppMetadata.authorEmail)")!)
                    .font(.caption).foregroundStyle(.secondary)
                Link("LinkedIn", destination: URL(string: AppMetadata.authorLinkedin)!)
                    .font(.caption)
            }

            Spacer()

            // Links
            HStack(spacing: 24) {
                Link(destination: URL(string: AppMetadata.website)!) {
                    Label("Website", systemImage: "globe")
                }
                Link(destination: URL(string: AppMetadata.repository)!) {
                    Label("GitHub", systemImage: "chevron.left.forwardslash.chevron.right")
                }
            }
            .font(.callout)

            Spacer()
        }
        .padding(.horizontal, 32)
        .padding(.vertical, 20)
    }

    private var footer: some View {
        Text(AppMetadata.copyright)
            .font(.caption2)
            .foregroundStyle(.tertiary)
            .frame(maxWidth: .infinity)
            .padding(.vertical, 12)
    }
}

#Preview {
    AboutView()
}
